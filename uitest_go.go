package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
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
		chromedp.Navigate("https://regulation.eaeunion.org/pd/"),
		chromedp.WaitVisible(".DocSearchResult_Item:nth-child(20)"),
		chromedp.Nodes(".DocSearchResult_Item", &baseNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Number of basenodes:", len(baseNodes))

	red := regexp.MustCompile(`\d{2}.\d{2}.\d{4}`)

	var dept, status, dates, etap string
	i := 0
	for _, node := range baseNodes {
		i++
		err = chromedp.Run(ctx,
			chromedp.Text(".DocSearchResult_Item__Date", &status, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(".DocSearchResult_Item__DatesLeft > div:nth-child(2)", &dept, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(".DocSearchResult_Item__DatesRight > div:nth-child(1)", &dates, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(".DocSearchResult_Item__DatesRight > div:nth-child(2)", &etap, chromedp.ByQuery, chromedp.FromNode(node)),
		)
		if err != nil {
			log.Fatal("Error:", err)
		}
		// Возможно два этапа, если что-то иное, то ошибка
		if !(strings.Trim(status, "\n\r") == "Общественное обсуждение" || strings.Trim(status, "\n\r") == "Оценка регулирующего воздействия") {
			fmt.Println("Неизвестный статус:", status, " в позиции - ", i)
		}
		// Если наименование департамента пустое (длина пусть будет меньше 10 байт), то ошибка
		if len(dept) < 10 {
			fmt.Println("Департамент не указан в позиции - ", i)
		}
		flist := red.FindAllString(dates, -1)
		sstartdate := strings.Split(flist[0], ".")
		senddate := strings.Split(flist[1], ".")
		var td, startdate, enddate time.Time
		startdate, _ = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", sstartdate[2], sstartdate[1], sstartdate[0]))
		enddate, _ = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", senddate[2], senddate[1], senddate[0]))
		today := time.Now()
		t_year, t_month, t_day := today.Date()
		td, _ = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", t_year, t_month, t_day))
		if td == startdate && !(strings.Trim(etap, "\n\r") == "Создан") {
			fmt.Println("Неверный этап:", etap, " в позиции - ", i)
		} else if td.After(startdate) && td.Before(enddate) && !(strings.Trim(etap, "\n\r") == "Идет обсуждение") {
			fmt.Println("Неверный этап:", etap, " в позиции - ", i)
		} else if td.After(enddate) && !(strings.Trim(etap, "\n\r") == "Обсуждение завершено") {
			fmt.Println("Неверный этап:", etap, " в позиции - ", i)
		}
	}
}
