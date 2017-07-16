package main

import (
	"fmt"

	"os"

	"github.com/arteev/go-translate"
	_ "github.com/arteev/go-yandex"
)

var apikey = os.Getenv("YANDEX_TRANSLATE_KEY")

func main() {
	tr, err := translate.New("yandex",
		translate.WithOption("apikey", apikey),
	)
	if err != nil {
		panic(err)
	}
	v, err := tr.GetLangs("ru")
	if err != nil {
		fmt.Println(err)
	} else {

		for _, l := range v {
			fmt.Println(l.Code, l.Name)
			fmt.Println("--->>>>")
			for _, to := range l.Dirs {
				fmt.Println("\t", to.Code, to.Name)
			}
		}
	}

	if l, err := tr.Detect("Перевод тест"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Detected: ", l)
	}

	if res := tr.Translate("Переведи меня", "en"); res.Err != nil {
		fmt.Println("Error:", res)
	} else {
		fmt.Println("Translate:", res.Text, " direction", res.FromLang, "-", res.ToLang, " detected:", res.Detected)
	}

}
