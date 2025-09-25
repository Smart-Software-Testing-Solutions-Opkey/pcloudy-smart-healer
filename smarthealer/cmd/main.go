package main

/*

#include <stdlib.h>
#include <stdint.h>

// Enums
typedef enum { Android = 0, Ios = 1, Web = 2 } Platform;
typedef enum { XMLPageType = 0, HTMLPageType = 1 } PageType;
typedef enum { Automatic = 0, Manual = 1, Screenshot = 2} ComparisionMode;

// API Structs
typedef struct {
	const char* project_id;
	const char* page_source;
	const char* b64_png;
	const char* xpath;
	const char* context_id;
	Platform platform;
	PageType page_type;
} Info;

typedef struct {
	ComparisionMode comparisionMode;
} Options;

typedef struct {
	const char* openai_key;
	const char* sqlite_db_path;
} Config;

typedef struct {
	int success;
	const char* reason;
	const char* content;
} Result;
*/
import "C"
import (
	"unsafe"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/healer"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/retrieval"
)

var (
	sh *smarthealer.SmartHealer
)

//export initSmartHealer
func initSmartHealer(conf C.Config) C.Result {
	cfg := convertConfig(conf)

	var err error
	sh, err = smarthealer.NewSmartHealer(cfg)
	if err != nil {
		return makeResult(false, err.Error(), "")
	}

	sh.StartBackgroundWorkers()

	return makeResult(true, "", "")
}

//export resolveLocator
func resolveLocator(cinfo C.Info, copt C.Options) C.Result {
	info := convertInfo(cinfo)
	opt := convertOpts(copt)

	l, err := sh.ResolveLocator(info, opt)
	if err != nil {
		return makeResult(false, err.Error(), "")
	}
	return makeResult(true, "", l)
}

//export resolveLocatorAsync
func resolveLocatorAsync(cinfo C.Info, copt C.Options) C.Result {
	info := convertInfo(cinfo)
	opt := convertOpts(copt)

	err := sh.ResolveLocatorAsync(info, opt)
	if err != nil {
		return makeResult(false, err.Error(), "")
	}
	return makeResult(true, "", "")
}

//export close
func close() {
	sh.Close()
}

//export freeResult
func freeResult(r C.Result) {
	if r.reason != nil {
		C.free(unsafe.Pointer(r.reason))
	}
	if r.content != nil {
		C.free(unsafe.Pointer(r.content))
	}
}

func goStringToC(s string) *C.char {
	if s == "" {
		return nil
	}
	return C.CString(s)
}

func cStringToGo(cstr *C.char) string {
	if cstr == nil {
		return ""
	}
	return C.GoString(cstr)
}

func makeResult(success bool, reason, content string) C.Result {
	var r C.Result

	s := 0
	if success {
		s = 1
	}

	r.success = C.int(s)
	r.reason = goStringToC(reason)
	r.content = goStringToC(content)

	return r
}

func convertConfig(conf C.Config) config.Config {
	cfg := config.Config{
		Db: config.DbConfig{
			Path: cStringToGo(conf.sqlite_db_path),
		},
		Ai: config.OpenAIConfig{
			SecretKey: cStringToGo(conf.openai_key),
		},
	}

	return cfg
}

func convertInfo(info C.Info) healer.LocatorInfo {
	i := healer.LocatorInfo{
		ProjectId:  cStringToGo(info.project_id),
		PageSource: cStringToGo(info.page_source),
		B64Png:     cStringToGo(info.b64_png),
		XPath:      cStringToGo(info.xpath),
		ContextId:  cStringToGo(info.context_id),
		Platform:   platform.Platform(int(info.platform)),
		PageType:   page.PageType(int(info.page_type)),
	}
	return i
}

func convertOpts(opt C.Options) healer.ResolveOptions {
	o := healer.ResolveOptions{
		ComparisionMode: retrieval.ComparisionMode(int(opt.comparisionMode)),
	}
	return o
}

// is Required
func main() {

}
