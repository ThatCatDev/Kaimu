package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper: creates org + project, returns projectID, boardID, and column name->ID map.
func bulkSetupProject(t *testing.T, server *BoardTestServer, token, name, key string) (projectID, boardID string, columns map[string]string) {
	createOrgQuery := fmt.Sprintf(`mutation { createOrganization(input: { name: "%s Org" }) { id } }`, name)
	orgResp := server.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors, "Create org errors: %v", orgResp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "%s", key: "%s" }) {
			id
			defaultBoard {
				id
				columns { id name }
			}
		}
	}`, orgData.CreateOrganization.ID, name, key)

	projResp := server.executeQuery(createProjectQuery, token)
	require.Empty(t, projResp.Errors, "Create project errors: %v", projResp.Errors)

	var projData struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)

	columns = make(map[string]string)
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		columns[col.Name] = col.ID
	}

	return projData.CreateProject.ID, projData.CreateProject.DefaultBoard.ID, columns
}

// Helper: creates a card in the given column, returns the card ID.
func bulkCreateCard(t *testing.T, server *BoardTestServer, token, columnID, title string) string {
	query := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "%s" }) { id }
	}`, columnID, title)

	resp := server.executeQuery(query, token)
	require.Empty(t, resp.Errors, "Create card '%s' errors: %v", title, resp.Errors)

	var data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(resp.Data, &data)
	require.NotEmpty(t, data.CreateCard.ID)
	return data.CreateCard.ID
}

// Helper: creates a tag in the given project, returns the tag ID.
func bulkCreateTag(t *testing.T, server *BoardTestServer, token, projectID, name, color string) string {
	query := fmt.Sprintf(`mutation {
		createTag(input: { projectId: "%s", name: "%s", color: "%s" }) { id }
	}`, projectID, name, color)

	resp := server.executeQuery(query, token)
	require.Empty(t, resp.Errors, "Create tag '%s' errors: %v", name, resp.Errors)

	var data struct {
		CreateTag struct {
			ID string `json:"id"`
		} `json:"createTag"`
	}
	json.Unmarshal(resp.Data, &data)
	require.NotEmpty(t, data.CreateTag.ID)
	return data.CreateTag.ID
}

func TestBulkMoveCardsToColumn(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("bulkmoveuser", "password123")
	require.NoError(t, err)

	_, boardID, columns := bulkSetupProject(t, server, token, "Bulk Move Test", "BMT")

	todoColID := columns["Todo"]
	inProgressColID := columns["In Progress"]
	require.NotEmpty(t, todoColID)
	require.NotEmpty(t, inProgressColID)

	// Create 3 cards in Todo
	card1ID := bulkCreateCard(t, server, token, todoColID, "Bulk Move Card 1")
	card2ID := bulkCreateCard(t, server, token, todoColID, "Bulk Move Card 2")
	card3ID := bulkCreateCard(t, server, token, todoColID, "Bulk Move Card 3")

	// Bulk move all 3 cards to In Progress
	bulkMoveQuery := fmt.Sprintf(`mutation {
		bulkMoveCardsToColumn(input: {
			cardIds: ["%s", "%s", "%s"]
			targetColumnId: "%s"
			boardId: "%s"
		}) {
			successCount
			totalCount
			cards { id }
		}
	}`, card1ID, card2ID, card3ID, inProgressColID, boardID)

	moveResp := server.executeQuery(bulkMoveQuery, token)
	require.Empty(t, moveResp.Errors, "Bulk move errors: %v", moveResp.Errors)

	var moveData struct {
		BulkMoveCardsToColumn struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID string `json:"id"`
			} `json:"cards"`
		} `json:"bulkMoveCardsToColumn"`
	}
	json.Unmarshal(moveResp.Data, &moveData)

	assert.Equal(t, 3, moveData.BulkMoveCardsToColumn.SuccessCount)
	assert.Equal(t, 3, moveData.BulkMoveCardsToColumn.TotalCount)
	assert.Equal(t, 3, len(moveData.BulkMoveCardsToColumn.Cards))

	// Verify each card is now in the In Progress column
	for _, cardID := range []string{card1ID, card2ID, card3ID} {
		queryCard := fmt.Sprintf(`query { card(id: "%s") { id column { id name } } }`, cardID)
		cardResp := server.executeQuery(queryCard, token)
		require.Empty(t, cardResp.Errors)

		var cardData struct {
			Card struct {
				ID     string `json:"id"`
				Column struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"column"`
			} `json:"card"`
		}
		json.Unmarshal(cardResp.Data, &cardData)
		assert.Equal(t, "In Progress", cardData.Card.Column.Name, "Card %s should be in In Progress", cardID)
	}
}

func TestBulkUpdateCardProperties(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("bulkpropsuser", "password123")
	require.NoError(t, err)

	_, boardID, columns := bulkSetupProject(t, server, token, "Bulk Props Test", "BPT")

	todoColID := columns["Todo"]
	require.NotEmpty(t, todoColID)

	// Create 2 cards
	card1ID := bulkCreateCard(t, server, token, todoColID, "Props Card 1")
	card2ID := bulkCreateCard(t, server, token, todoColID, "Props Card 2")

	// --- Test 1: Bulk set priority to HIGH ---
	bulkPriorityQuery := fmt.Sprintf(`mutation {
		bulkUpdateCardProperties(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
			priority: HIGH
		}) {
			successCount
			totalCount
			cards { id priority }
		}
	}`, card1ID, card2ID, boardID)

	prioResp := server.executeQuery(bulkPriorityQuery, token)
	require.Empty(t, prioResp.Errors, "Bulk priority errors: %v", prioResp.Errors)

	var prioData struct {
		BulkUpdateCardProperties struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID       string `json:"id"`
				Priority string `json:"priority"`
			} `json:"cards"`
		} `json:"bulkUpdateCardProperties"`
	}
	json.Unmarshal(prioResp.Data, &prioData)

	assert.Equal(t, 2, prioData.BulkUpdateCardProperties.SuccessCount)
	assert.Equal(t, 2, prioData.BulkUpdateCardProperties.TotalCount)
	for _, c := range prioData.BulkUpdateCardProperties.Cards {
		assert.Equal(t, "HIGH", c.Priority, "Card %s should have priority HIGH", c.ID)
	}

	// --- Test 2: Bulk set assignee ---
	// Get the current user's ID via the me query
	meResp := server.executeQuery(`query { me { id } }`, token)
	require.Empty(t, meResp.Errors, "me query errors: %v", meResp.Errors)

	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(meResp.Data, &meData)
	userID := meData.Me.ID
	require.NotEmpty(t, userID)

	bulkAssignQuery := fmt.Sprintf(`mutation {
		bulkUpdateCardProperties(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
			assigneeId: "%s"
		}) {
			successCount
			totalCount
			cards { id assignee { id } }
		}
	}`, card1ID, card2ID, boardID, userID)

	assignResp := server.executeQuery(bulkAssignQuery, token)
	require.Empty(t, assignResp.Errors, "Bulk assignee errors: %v", assignResp.Errors)

	var assignData struct {
		BulkUpdateCardProperties struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID       string `json:"id"`
				Assignee *struct {
					ID string `json:"id"`
				} `json:"assignee"`
			} `json:"cards"`
		} `json:"bulkUpdateCardProperties"`
	}
	json.Unmarshal(assignResp.Data, &assignData)

	assert.Equal(t, 2, assignData.BulkUpdateCardProperties.SuccessCount)
	for _, c := range assignData.BulkUpdateCardProperties.Cards {
		require.NotNil(t, c.Assignee, "Card %s should have an assignee", c.ID)
		assert.Equal(t, userID, c.Assignee.ID, "Card %s assignee should be %s", c.ID, userID)
	}

	// --- Test 3: Bulk clear assignee ---
	bulkClearAssignQuery := fmt.Sprintf(`mutation {
		bulkUpdateCardProperties(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
			clearAssignee: true
		}) {
			successCount
			totalCount
			cards { id assignee { id } }
		}
	}`, card1ID, card2ID, boardID)

	clearResp := server.executeQuery(bulkClearAssignQuery, token)
	require.Empty(t, clearResp.Errors, "Bulk clear assignee errors: %v", clearResp.Errors)

	var clearData struct {
		BulkUpdateCardProperties struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID       string `json:"id"`
				Assignee *struct {
					ID string `json:"id"`
				} `json:"assignee"`
			} `json:"cards"`
		} `json:"bulkUpdateCardProperties"`
	}
	json.Unmarshal(clearResp.Data, &clearData)

	assert.Equal(t, 2, clearData.BulkUpdateCardProperties.SuccessCount)
	for _, c := range clearData.BulkUpdateCardProperties.Cards {
		assert.Nil(t, c.Assignee, "Card %s should have no assignee after clear", c.ID)
	}
}

func TestBulkTagCards(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("bulktaguser", "password123")
	require.NoError(t, err)

	projectID, boardID, columns := bulkSetupProject(t, server, token, "Bulk Tag Test", "BTT")

	todoColID := columns["Todo"]
	require.NotEmpty(t, todoColID)

	// Create 2 cards
	card1ID := bulkCreateCard(t, server, token, todoColID, "Tag Card 1")
	card2ID := bulkCreateCard(t, server, token, todoColID, "Tag Card 2")

	// Create 2 tags
	tag1ID := bulkCreateTag(t, server, token, projectID, "Bug", "#EF4444")
	tag2ID := bulkCreateTag(t, server, token, projectID, "Feature", "#10B981")

	// --- Test 1: Bulk ADD both tags to both cards ---
	bulkAddTagsQuery := fmt.Sprintf(`mutation {
		bulkTagCards(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
			tagIds: ["%s", "%s"]
			operation: ADD
		}) {
			successCount
			totalCount
			cards { id tags { id name } }
		}
	}`, card1ID, card2ID, boardID, tag1ID, tag2ID)

	addResp := server.executeQuery(bulkAddTagsQuery, token)
	require.Empty(t, addResp.Errors, "Bulk add tags errors: %v", addResp.Errors)

	var addData struct {
		BulkTagCards struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID   string `json:"id"`
				Tags []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"tags"`
			} `json:"cards"`
		} `json:"bulkTagCards"`
	}
	json.Unmarshal(addResp.Data, &addData)

	assert.Equal(t, 2, addData.BulkTagCards.SuccessCount)
	assert.Equal(t, 2, addData.BulkTagCards.TotalCount)
	for _, c := range addData.BulkTagCards.Cards {
		assert.Equal(t, 2, len(c.Tags), "Card %s should have 2 tags", c.ID)
	}

	// --- Test 2: Bulk REMOVE one tag from both cards ---
	bulkRemoveTagQuery := fmt.Sprintf(`mutation {
		bulkTagCards(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
			tagIds: ["%s"]
			operation: REMOVE
		}) {
			successCount
			totalCount
			cards { id tags { id name } }
		}
	}`, card1ID, card2ID, boardID, tag1ID)

	removeResp := server.executeQuery(bulkRemoveTagQuery, token)
	require.Empty(t, removeResp.Errors, "Bulk remove tag errors: %v", removeResp.Errors)

	var removeData struct {
		BulkTagCards struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID   string `json:"id"`
				Tags []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"tags"`
			} `json:"cards"`
		} `json:"bulkTagCards"`
	}
	json.Unmarshal(removeResp.Data, &removeData)

	assert.Equal(t, 2, removeData.BulkTagCards.SuccessCount)
	for _, c := range removeData.BulkTagCards.Cards {
		assert.Equal(t, 1, len(c.Tags), "Card %s should have 1 tag after removal", c.ID)
		assert.Equal(t, "Feature", c.Tags[0].Name, "Remaining tag should be Feature")
	}
}

func TestBulkDeleteCards(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("bulkdeluser", "password123")
	require.NoError(t, err)

	_, boardID, columns := bulkSetupProject(t, server, token, "Bulk Delete Test", "BDT")

	todoColID := columns["Todo"]
	require.NotEmpty(t, todoColID)

	// Create 3 cards
	card1ID := bulkCreateCard(t, server, token, todoColID, "Delete Card 1")
	card2ID := bulkCreateCard(t, server, token, todoColID, "Delete Card 2")
	card3ID := bulkCreateCard(t, server, token, todoColID, "Delete Card 3")

	// Bulk delete cards 1 and 2
	bulkDeleteQuery := fmt.Sprintf(`mutation {
		bulkDeleteCards(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
		})
	}`, card1ID, card2ID, boardID)

	deleteResp := server.executeQuery(bulkDeleteQuery, token)
	require.Empty(t, deleteResp.Errors, "Bulk delete errors: %v", deleteResp.Errors)

	var deleteData struct {
		BulkDeleteCards bool `json:"bulkDeleteCards"`
	}
	json.Unmarshal(deleteResp.Data, &deleteData)
	assert.True(t, deleteData.BulkDeleteCards)

	// Verify card 1 is gone
	query1 := fmt.Sprintf(`query { card(id: "%s") { id } }`, card1ID)
	resp1 := server.executeQuery(query1, token)
	// Deleted card should return an error or null
	var card1Data struct {
		Card *struct {
			ID string `json:"id"`
		} `json:"card"`
	}
	json.Unmarshal(resp1.Data, &card1Data)
	assert.Nil(t, card1Data.Card, "Card 1 should be deleted")

	// Verify card 2 is gone
	query2 := fmt.Sprintf(`query { card(id: "%s") { id } }`, card2ID)
	resp2 := server.executeQuery(query2, token)
	var card2Data struct {
		Card *struct {
			ID string `json:"id"`
		} `json:"card"`
	}
	json.Unmarshal(resp2.Data, &card2Data)
	assert.Nil(t, card2Data.Card, "Card 2 should be deleted")

	// Verify card 3 still exists
	query3 := fmt.Sprintf(`query { card(id: "%s") { id title } }`, card3ID)
	resp3 := server.executeQuery(query3, token)
	require.Empty(t, resp3.Errors)
	var card3Data struct {
		Card *struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"card"`
	}
	json.Unmarshal(resp3.Data, &card3Data)
	require.NotNil(t, card3Data.Card, "Card 3 should still exist")
	assert.Equal(t, "Delete Card 3", card3Data.Card.Title)
}

func TestBulkMoveCardsToBacklog(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("bulkbackloguser", "password123")
	require.NoError(t, err)

	_, boardID, columns := bulkSetupProject(t, server, token, "Bulk Backlog Test", "BBT")

	todoColID := columns["Todo"]
	require.NotEmpty(t, todoColID)

	// Create a sprint
	startDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 13).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Bulk Backlog Sprint"
			startDate: "%s"
			endDate: "%s"
		}) { id name }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	require.Empty(t, sprintResp.Errors, "Create sprint errors: %v", sprintResp.Errors)

	var sprintData struct {
		CreateSprint struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID
	require.NotEmpty(t, sprintID)

	// Start the sprint
	startSprintQuery := fmt.Sprintf(`mutation { startSprint(id: "%s") { id status } }`, sprintID)
	startResp := server.executeQuery(startSprintQuery, token)
	require.Empty(t, startResp.Errors, "Start sprint errors: %v", startResp.Errors)

	var startData struct {
		StartSprint struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"startSprint"`
	}
	json.Unmarshal(startResp.Data, &startData)
	assert.Equal(t, "ACTIVE", startData.StartSprint.Status)

	// Create 3 cards
	card1ID := bulkCreateCard(t, server, token, todoColID, "Backlog Card 1")
	card2ID := bulkCreateCard(t, server, token, todoColID, "Backlog Card 2")
	card3ID := bulkCreateCard(t, server, token, todoColID, "Backlog Card 3")

	// Add all 3 cards to the sprint
	for _, cardID := range []string{card1ID, card2ID, card3ID} {
		addQuery := fmt.Sprintf(`mutation {
			addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) {
				id
				sprints { id }
			}
		}`, cardID, sprintID)
		addResp := server.executeQuery(addQuery, token)
		require.Empty(t, addResp.Errors, "Add card %s to sprint errors: %v", cardID, addResp.Errors)
	}

	// Verify cards are in the sprint
	for _, cardID := range []string{card1ID, card2ID, card3ID} {
		getQuery := fmt.Sprintf(`query { card(id: "%s") { id sprints { id } } }`, cardID)
		getResp := server.executeQuery(getQuery, token)
		require.Empty(t, getResp.Errors)

		var getData struct {
			Card struct {
				Sprints []struct {
					ID string `json:"id"`
				} `json:"sprints"`
			} `json:"card"`
		}
		json.Unmarshal(getResp.Data, &getData)
		assert.Equal(t, 1, len(getData.Card.Sprints), "Card %s should be in 1 sprint", cardID)
	}

	// Bulk move cards 1 and 2 to backlog
	bulkBacklogQuery := fmt.Sprintf(`mutation {
		bulkMoveCardsToBacklog(input: {
			cardIds: ["%s", "%s"]
			boardId: "%s"
		}) {
			successCount
			totalCount
			cards { id }
		}
	}`, card1ID, card2ID, boardID)

	backlogResp := server.executeQuery(bulkBacklogQuery, token)
	require.Empty(t, backlogResp.Errors, "Bulk move to backlog errors: %v", backlogResp.Errors)

	var backlogData struct {
		BulkMoveCardsToBacklog struct {
			SuccessCount int `json:"successCount"`
			TotalCount   int `json:"totalCount"`
			Cards        []struct {
				ID string `json:"id"`
			} `json:"cards"`
		} `json:"bulkMoveCardsToBacklog"`
	}
	json.Unmarshal(backlogResp.Data, &backlogData)

	assert.Equal(t, 2, backlogData.BulkMoveCardsToBacklog.SuccessCount)
	assert.Equal(t, 2, backlogData.BulkMoveCardsToBacklog.TotalCount)
	assert.Equal(t, 2, len(backlogData.BulkMoveCardsToBacklog.Cards))

	// Verify cards 1 and 2 are no longer in any sprint
	for _, cardID := range []string{card1ID, card2ID} {
		getQuery := fmt.Sprintf(`query { card(id: "%s") { id sprints { id } } }`, cardID)
		getResp := server.executeQuery(getQuery, token)
		require.Empty(t, getResp.Errors)

		var getData struct {
			Card struct {
				Sprints []struct {
					ID string `json:"id"`
				} `json:"sprints"`
			} `json:"card"`
		}
		json.Unmarshal(getResp.Data, &getData)
		assert.Equal(t, 0, len(getData.Card.Sprints), "Card %s should have no sprints after backlog move", cardID)
	}

	// Verify card 3 is still in the sprint
	getCard3Query := fmt.Sprintf(`query { card(id: "%s") { id sprints { id } } }`, card3ID)
	getCard3Resp := server.executeQuery(getCard3Query, token)
	require.Empty(t, getCard3Resp.Errors)

	var getCard3Data struct {
		Card struct {
			Sprints []struct {
				ID string `json:"id"`
			} `json:"sprints"`
		} `json:"card"`
	}
	json.Unmarshal(getCard3Resp.Data, &getCard3Data)
	assert.Equal(t, 1, len(getCard3Data.Card.Sprints), "Card 3 should still be in 1 sprint")
}
