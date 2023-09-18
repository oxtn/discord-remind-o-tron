package commandhandlers

import (
	"discord-remind-o-tron/util"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func Remind(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	util.LogJson(data)
	switch data.Name {
	case "remind":
		err := s.InteractionRespond(
			i.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "K",
				},
			},
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}
