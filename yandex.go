package yandex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/arteev/go-translate"
	"github.com/arteev/go-translate/translator"
)

const (
	PROVIDER_CODE   = "yandex"
	URL             = "https://translate.yandex.net/api/v1.5/tr.json"
	ROUTE_TRANSLATE = "translate"
	ROUTE_DETECT    = "detect"
	ROUTE_LANGS     = "getLangs"
)

//todo: yandex test!

type ProviderYandex struct {
	apikey string
}

type response struct {
	Code    int
	Message string
}

//response api getLangs
type languages struct {
	response
	Dirs  []string
	Langs map[string]string
}

//response api detect
type respdetect struct {
	response
	Lang string
}

//response api translate
type responseTranslate struct {
	response
	Lang     string
	Text     []string
	Detected map[string]string
}

func urlroute(route string) string {
	return URL + "/" + route
}

func (p *ProviderYandex) GetLangs(code string) ([]*translator.Language, error) {
	r, err := http.PostForm(urlroute(ROUTE_LANGS), url.Values{
		"key": {p.apikey},
		"ui":  {code},
	})
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var langs languages
	if err := json.NewDecoder(r.Body).Decode(&langs); err != nil {
		return nil, err
	}
	if err := p.DecodeApiError(langs.Code, 0, langs.Message); err != nil {
		return nil, err
	}
	type suplangs struct {
		l    *translator.Language
		from bool
	}
	mlangs := make(map[string]*suplangs)
	result := make([]*translator.Language, 0)
	for _, dir := range langs.Dirs {
		fromto := strings.Split(dir, "-")
		var curlang, curname, to, toname string
		if len(fromto) <= 1 {
			curlang = dir
		} else {
			curlang = fromto[0]
			to = fromto[1]
		}
		if langs.Langs != nil {
			curname, _ = langs.Langs[curlang]
			toname, _ = langs.Langs[to]
		}

		var lfrom, lto *suplangs
		var ok bool
		lfrom, ok = mlangs[curlang]
		if !ok {
			lfrom = &suplangs{translator.NewLanguage(curlang, curname), true}
			mlangs[curlang] = lfrom
			result = append(result, lfrom.l)
		} else if !lfrom.from {
			lfrom.from = true
			result = append(result, lfrom.l)
		}
		lto, ok = mlangs[to]
		if !ok {
			lto = &suplangs{translator.NewLanguage(to, toname), false}
			mlangs[to] = lto
		}
		lfrom.l.AddDir(lto.l)
	}
	return result, nil
}

func (p *ProviderYandex) Detect(text string) (*translator.Language, error) {
	r, err := http.PostForm(urlroute(ROUTE_DETECT), url.Values{
		"key":  {p.apikey},
		"text": {text},
	})
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var response respdetect
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}
	if err := p.DecodeApiError(response.Code, 200, response.Message); err != nil {
		return nil, err
	}
	return translator.NewLanguage(response.Lang, ""), nil
}

func (p *ProviderYandex) Translate(text, direction string) *translator.Result {
	r, err := http.PostForm(urlroute(ROUTE_TRANSLATE), url.Values{
		"key":     {p.apikey},
		"text":    {text},
		"lang":    {direction},
		"format":  {"plain"},
		"options": {"1"},
	})
	if err != nil {
		return &translator.Result{Err: err}
	}
	defer r.Body.Close()
	var response responseTranslate
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return &translator.Result{Err: err}
	}
	if err := p.DecodeApiError(response.Code, 200, response.Message); err != nil {
		return &translator.Result{Err: err}
	}

	result := &translator.Result{}
	result.Text = response.Text[0]
	if l, ok := response.Detected["lang"]; ok {
		result.Detected = translator.NewLanguage(l, "")
	}

	fromto := strings.Split(response.Lang, "-")
	if len(fromto) <= 1 {
		result.FromLang = translator.NewLanguage(response.Lang, "")
	} else {
		result.FromLang = translator.NewLanguage(fromto[0], "")
		result.ToLang = translator.NewLanguage(fromto[1], "")
	}

	return result

}

func (p *ProviderYandex) Name() string {
	return PROVIDER_CODE
}

func (p *ProviderYandex) DecodeApiError(code, success int, message string) error {
	if code == success {
		return nil
	}
	switch code {
	case 401:
		return translator.ErrWrongApiKey
	case 402:
		return translator.ErrBlockedApiKey
	case 403:
		return translator.ErrLimitDayExceeded
	case 404:
		return translator.ErrLimitMonthExceeded
	case 413:
		return translator.ErrLimitTextExceeded
	case 422:
		return translator.ErrTextNotTranslated
	case 501:
		return translator.ErrDirectionUnsupported
	default:
		return fmt.Errorf("Bad request: (%d) %s", code, message)
	}

}

type transfact struct{}

func (transfact) NewInstance(opts map[string]interface{}) translator.Translator {
	result := &ProviderYandex{}
	//TODO: check apikey
	result.apikey = opts["apikey"].(string)
	return translator.Translator(result)
}

func init() {
	translate.Register(PROVIDER_CODE, &transfact{})
}
