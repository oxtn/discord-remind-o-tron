package commandhandlers

import (
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/bwmarrin/discordgo"
)

// Mocks and helper functions for testing
type fakeSession struct {
	*discordgo.Session
}

func (s *fakeSession) ChannelMessageSend(m string, c string) (*discordgo.Message, error) {
	return nil, nil
}

func createFakeInteractionCreate() *discordgo.InteractionCreate {
	data := discordgo.ApplicationCommandInteractionData{
		Name: "remind",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString,
				Name:  "when",
				Value: "2023-10-05T12:00:00Z", // A future date
			},
			{
				Type:  discordgo.ApplicationCommandOptionUser,
				Name:  "target",
				Value: gofakeit.UUID(),
			},
			{
				Type: discordgo.ApplicationCommandOptionString,
				Name: "message",

				Value: "Test Reminder",
			},
		},
	}
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Data: &data,
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       gofakeit.UUID(),
					Username: gofakeit.Username(),
				},
			},
		},
	}
}

func TestHandler_ValidInput(t *testing.T) {
	remind := &Remind{} // Initialize your Remind struct here
	s := &fakeSession{Session: &discordgo.Session{}}
	i := createFakeInteractionCreate()

	// Call the Handler function
	remind.Handler(s, i)

	// Add assertions here to check the expected outcomes
	// Example:
	// assert.NoError(t, err) // Ensure no errors occurred during the test
}

func TestHandler_InvalidTimeFormat(t *testing.T) {
	remind := &Remind{} // Initialize your Remind struct here
	s := &fakeSession{Session: &discordgo.Session{}}
	i := createFakeInteractionCreate()
	i.ApplicationCommandData().Options[0].Value = "invalid-date-format"

	// Call the Handler function
	remind.Handler(s, i)

	// Add assertions here to check the expected outcomes for invalid time format
	// Example:
	// assert.Contains(t, "Your provided time of", response) // Ensure the error message is as expected
}

func TestHandler_PastTime(t *testing.T) {
	remind := &Remind{} // Initialize your Remind struct here
	s := &fakeSession{Session: &discordgo.Session{}}
	i := createFakeInteractionCreate()
	i.ApplicationCommandData().Options[0].Value = time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	// Call the Handler function
	remind.Handler(s, i)

	// Add assertions here to check the expected outcomes for a past time
	// Example:
	// assert.Contains(t, "Your provided time of", response) // Ensure the error message is as expected
}

// Add more test cases as needed to cover other scenarios

func TestMain(m *testing.M) {
	// Setup code here, e.g., initialize your database connection or other dependencies

	code := m.Run()

	// Teardown code here, e.g., close the database connection

	os.Exit(code)
}
