package whatsapp

import (
	"strings"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name         string
		msg          *waE2E.Message
		wantText     string
		wantContains string
	}{
		{
			name:     "nil message",
			msg:      nil,
			wantText: "",
		},
		{
			name:     "empty message",
			msg:      &waE2E.Message{},
			wantText: "",
		},
		{
			name:     "plain text",
			msg:      &waE2E.Message{Conversation: proto.String("hello world")},
			wantText: "hello world",
		},
		{
			name: "extended text",
			msg: &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String("rich text with link"),
			}},
			wantText: "rich text with link",
		},
		{
			name: "image with caption",
			msg: &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
				Caption: proto.String("photo caption"),
			}},
			wantText: "photo caption",
		},
		{
			name:     "image without caption",
			msg:      &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}},
			wantText: "",
		},
		{
			name: "video with caption",
			msg: &waE2E.Message{VideoMessage: &waE2E.VideoMessage{
				Caption: proto.String("video caption"),
			}},
			wantText: "video caption",
		},
		{
			name: "document with caption",
			msg: &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
				Caption: proto.String("doc caption"),
			}},
			wantText: "doc caption",
		},
		{
			name: "document with title only",
			msg: &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
				Title: proto.String("report.pdf"),
			}},
			wantText: "report.pdf",
		},
		{
			name:     "sticker",
			msg:      &waE2E.Message{StickerMessage: &waE2E.StickerMessage{}},
			wantText: "[Sticker]",
		},
		{
			name: "location with name",
			msg: &waE2E.Message{LocationMessage: &waE2E.LocationMessage{
				Name:             proto.String("Raffles Place"),
				DegreesLatitude:  proto.Float64(1.2840),
				DegreesLongitude: proto.Float64(103.8510),
			}},
			wantText: "[Location: Raffles Place]",
		},
		{
			name: "location without name",
			msg: &waE2E.Message{LocationMessage: &waE2E.LocationMessage{
				DegreesLatitude:  proto.Float64(1.2840),
				DegreesLongitude: proto.Float64(103.8510),
			}},
			wantContains: "[Location:",
		},
		{
			name: "contact card",
			msg: &waE2E.Message{ContactMessage: &waE2E.ContactMessage{
				DisplayName: proto.String("John Doe"),
			}},
			wantText: "[Contact: John Doe]",
		},
		{
			name: "poll",
			msg: &waE2E.Message{PollCreationMessage: &waE2E.PollCreationMessage{
				Name: proto.String("Favourite food?"),
			}},
			wantText: "[Poll: Favourite food?]",
		},
		{
			name: "reaction",
			msg: &waE2E.Message{ReactionMessage: &waE2E.ReactionMessage{
				Text: proto.String("👍"),
			}},
			wantText: "[Reaction: 👍]",
		},
		{
			name: "hydrated four row template (business authentication code)",
			msg: &waE2E.Message{TemplateMessage: &waE2E.TemplateMessage{
				Format: &waE2E.TemplateMessage_HydratedFourRowTemplate_{HydratedFourRowTemplate: &waE2E.TemplateMessage_HydratedFourRowTemplate{
					Title:               &waE2E.TemplateMessage_HydratedFourRowTemplate_HydratedTitleText{HydratedTitleText: "Meta"},
					HydratedContentText: proto.String("123456 is your verification code."),
					HydratedFooterText:  proto.String("This code expires in 10 minutes."),
				}},
			}},
			wantText: "Meta\n123456 is your verification code.\nThis code expires in 10 minutes.",
		},
		{
			name: "template with legacy hydrated template field",
			msg: &waE2E.Message{TemplateMessage: &waE2E.TemplateMessage{
				HydratedTemplate: &waE2E.TemplateMessage_HydratedFourRowTemplate{
					HydratedContentText: proto.String("654321 is your code."),
				},
			}},
			wantText: "654321 is your code.",
		},
		{
			name: "interactive message with copy code button",
			msg: &waE2E.Message{InteractiveMessage: &waE2E.InteractiveMessage{
				Header: &waE2E.InteractiveMessage_Header{Title: proto.String("Verify")},
				Body:   &waE2E.InteractiveMessage_Body{Text: proto.String("Your code is 112233")},
				Footer: &waE2E.InteractiveMessage_Footer{Text: proto.String("Do not share it")},
			}},
			wantText: "Verify\nYour code is 112233\nDo not share it",
		},
		{
			name: "template wrapping an interactive message",
			msg: &waE2E.Message{TemplateMessage: &waE2E.TemplateMessage{
				Format: &waE2E.TemplateMessage_InteractiveMessageTemplate{InteractiveMessageTemplate: &waE2E.InteractiveMessage{
					Body: &waE2E.InteractiveMessage_Body{Text: proto.String("Code 445566")},
				}},
			}},
			wantText: "Code 445566",
		},
		{
			name: "buttons message",
			msg: &waE2E.Message{ButtonsMessage: &waE2E.ButtonsMessage{
				Header:      &waE2E.ButtonsMessage_Text{Text: "Header"},
				ContentText: proto.String("Pick one"),
				FooterText:  proto.String("Footer"),
			}},
			wantText: "Header\nPick one\nFooter",
		},
		{
			name: "list message",
			msg: &waE2E.Message{ListMessage: &waE2E.ListMessage{
				Title:       proto.String("Menu"),
				Description: proto.String("Choose an item"),
			}},
			wantText: "Menu\nChoose an item",
		},
		{
			name: "highly structured message with hydrated template",
			msg: &waE2E.Message{HighlyStructuredMessage: &waE2E.HighlyStructuredMessage{
				HydratedHsm: &waE2E.TemplateMessage{HydratedTemplate: &waE2E.TemplateMessage_HydratedFourRowTemplate{
					HydratedContentText: proto.String("778899 is your code."),
				}},
			}},
			wantText: "778899 is your code.",
		},
		{
			name: "buttons response",
			msg: &waE2E.Message{ButtonsResponseMessage: &waE2E.ButtonsResponseMessage{
				Response: &waE2E.ButtonsResponseMessage_SelectedDisplayText{SelectedDisplayText: "Yes"},
			}},
			wantText: "Yes",
		},
		{
			name: "template button reply",
			msg: &waE2E.Message{TemplateButtonReplyMessage: &waE2E.TemplateButtonReplyMessage{
				SelectedDisplayText: proto.String("Confirm"),
			}},
			wantText: "Confirm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTextContent(tt.msg)
			if tt.wantText != "" {
				if got != tt.wantText {
					t.Errorf("ExtractTextContent() = %q, want %q", got, tt.wantText)
				}
			} else if tt.wantContains != "" {
				if !strings.Contains(got, tt.wantContains) {
					t.Errorf("ExtractTextContent() = %q, want to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}
