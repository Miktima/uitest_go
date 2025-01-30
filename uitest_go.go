package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {

	type Configlist struct {
		Chrome string
		Uagent string
	}

	var config Configlist

	path, _ := os.Executable()
	path = path[:strings.LastIndex(path, "/")+1]

	//read config
	if _, err := os.Stat(path + "/config.json"); err == nil {
		// Open our jsonFile
		byteValue, err := os.ReadFile(path + "/config.json")
		// if we os.ReadFile returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(byteValue, &config)
		if err != nil {
			fmt.Println(err)
		}
	}

	// create context
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), config.Chrome)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// run task list
	chromedp.UserAgent(config.Uagent)

	// Проверка списка документов по общественным обсуждениям и оценкам регулирующего воздействия
	// на портале евразийского экономического союза

	var baseNodes []*cdp.Node
	// Получаем список документов
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://docs.eaeunion.org/ru-ru/pages/regulation.aspx`),
		chromedp.Sleep(2000*time.Millisecond),
		chromedp.WaitVisible(`.discussionsAndRIA-panel`, chromedp.ByQuery),
		chromedp.Nodes("//div[@class='discussionsAndRIA-panel']/table/tbody/tr", &baseNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Number of basenodes:", len(baseNodes))
	// В одну строку должны попадать два элемента <tr>
	ntr := 1
	var status, period string
	for _, node := range baseNodes {
		if ntr == 2 {
			err = chromedp.Run(ctx,
				chromedp.Text(".dd-status", &status, chromedp.ByQuery, chromedp.FromNode(node)),
				chromedp.Text(".dd-period", &period, chromedp.ByQuery, chromedp.FromNode(node)),
			)

			if err != nil {
				log.Fatal("Error:", err)
			}
			fmt.Println(period, "     ", status)
			ntr = 1
		} else {
			ntr = 2
		}
	}
}
