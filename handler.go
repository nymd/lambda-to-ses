package main

import (
	"log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	SenderMissing = errors.New("Missing sender")
	RecipientMissing = errors.New("Missing recipient")
	SubjectMissing = errors.New("Missing subject")
	TextMissing = errors.New("Missing body text")
	HTMLMissing = errors.New("Missing body HTML")
)

type InboundMessage struct {
	Sender string `json:"sender"`
	Recipient string `json:"recipient"`
	Subject string `json:"subject"`
	Text string `json:"text"`
	HTML string `json:"html"`
}

type ResponseMessage struct {
	Type string `json:"type"`
	Message string `json:"message"`
}

var emailClient *ses.SES

func init() {
	emailClient = ses.New(session.New(), aws.NewConfig().WithRegion("us-west-2"))
}

func ReturnErrorToUser(error error, status int) (events.APIGatewayProxyResponse, error) {
	errorMessage, _ := json.Marshal(ResponseMessage{
		Type: "error",
		Message: error.Error(),
	})

	log.Println(error.Error())

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{"Content-Type": "application/json"},
		StatusCode:status,
		Body: string(errorMessage),
	}, nil
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	body := request.Body

	var message InboundMessage
	err := json.Unmarshal([]byte(body), &message)

	if err != nil {
		return ReturnErrorToUser(err, http.StatusInternalServerError)
	} else if len(message.Sender) < 1 {
		return ReturnErrorToUser(SenderMissing, http.StatusBadRequest)
	} else if len(message.Recipient) < 1 {
		return ReturnErrorToUser(RecipientMissing, http.StatusBadRequest)
	}else if len(message.Subject) < 1 {
		return ReturnErrorToUser(SubjectMissing, http.StatusBadRequest)
	}else if len(message.Text) < 1 {
		return ReturnErrorToUser(TextMissing, http.StatusBadRequest)
	}else if len(message.HTML) < 1 {
		return ReturnErrorToUser(HTMLMissing, http.StatusBadRequest)
	}

	emailParams := &ses.SendEmailInput{
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data:aws.String(message.Text),
				},
				Html: &ses.Content{
					Data:aws.String(message.HTML),
				},
			},
			Subject: &ses.Content{
				Data:aws.String(message.Subject),
			},
		},
		Destination: &ses.Destination{
			ToAddresses:[]*string{aws.String(message.Recipient)},
		},
		Source:aws.String(message.Sender),
	}

	_, err = emailClient.SendEmail(emailParams)

	if err != nil {
		return ReturnErrorToUser(err, http.StatusInternalServerError)
	}

	successResponse, err := json.Marshal(ResponseMessage{"success", "Message is sent"})
	return events.APIGatewayProxyResponse{
		Body: string(successResponse),
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
