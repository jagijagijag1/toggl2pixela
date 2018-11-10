package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/gtoggl-api/gthttp"
	"github.com/dougEfresh/gtoggl-api/gttimentry"
	pixela "github.com/gainings/pixela-go-client"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) error {
	// extract env var
	apiToken := os.Getenv("TOGGL_API_TOKEN")
	pjID, _ := strconv.ParseUint(os.Getenv("TOGGL_PROJECT_ID"), 10, 64)
	user := os.Getenv("PIXELA_USER")
	token := os.Getenv("PIXELA_TOKEN")
	graph := os.Getenv("PIXELA_GRAPH")

	// extract data from toggl
	date, quantity := getDateAndTimeFromToggl(apiToken, pjID)
	if date == "-1" || quantity == "-1" {
		return errors.New("Error in accessing toggl")
	}
	fmt.Printf("date: %s, quantity: %s\n", date, quantity)

	// record pixel
	perr := recordPixel(user, token, graph, date, quantity)
	if perr != nil {
		return errors.New("Error in accessing pixela")
	}

	return nil
}

func getDateAndTimeFromToggl(apiToken string, pjID uint64) (string, string) {
	// create toggl client
	thc, err := gthttp.NewClient(apiToken)
	if err != nil {
		fmt.Println(err)
		return "-1", "-1"
	}

	// set time range to be analyzed
	y := time.Now().AddDate(0, 0, -1)
	s := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
	e := time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, 0, time.Local)
	date := y.Format("20060102")

	// get time entries
	total := int64(0)
	tec := gttimeentry.NewClient(thc)
	entries, eerr := tec.GetRange(s, e)
	if eerr != nil {
		fmt.Println(eerr)
		return "-1", "-1"
	}

	// sum durations with project pjID
	for _, e := range entries {
		if e.Pid == pjID {
			total += e.Duration
		}
	}
	totalMin := float64(total) / 60
	quantity := strconv.FormatFloat(totalMin, 'f', 4, 64)

	return date, quantity
}

func recordPixel(user, token, graph, date, quantity string) error {
	c := pixela.NewClient(user, token)

	// try to record
	err := c.RegisterPixel(graph, date, quantity)
	if err == nil {
		fmt.Println("recorded")
		return err
	}

	// if fail, try to update
	err = c.UpdatePixelQuantity(graph, date, quantity)
	if err == nil {
		fmt.Println("updated")
	}

	return err
}

func main() {
	lambda.Start(Handler)
}
