package toolbox

import (
	"errors"
	"fmt"
	"github.com/HouzuoGuo/laitos/browser/phantomjs"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// RegexBrowserCommandAndParam finds a browser command and an optional parameter.
	RegexBrowserCommandAndParam = regexp.MustCompile(`(\w+)[^\w]*(.*)?`)
	ErrBadBrowserParam          = errors.New(`(f/b,p/n/nn/0,r/k,g/i,ptr,val,e/enter/backsp) [param..]`)
)

// FormatElementInfoArray prints element information into strings.
func FormatElementInfoArrayPhantomJS(elements []phantomjs.ElementInfo) string {
	if elements == nil || len(elements) == 0 {
		return "(nothing)"
	}
	lines := make([]string, 0, len(elements))
	for _, elem := range elements {
		lines = append(lines, fmt.Sprintf("%s[%s]-\"%s\"-\"%v\"-%s##", elem.TagName, elem.ID, elem.Name, elem.Value, elem.InnerHTML))
	}
	return strings.Join(lines, "\n")
}

// BrowserPhantomJS offers remote control to exactly one PhantomJS page.
type BrowserPhantomJS struct {
	Renderers *phantomjs.Instances `json:"Browsers"` // Instances configure and manage phantomJS processes.
	renderer  *phantomjs.Instance  // renderer is the one and only browser instance tied to this feature
	mutex     *sync.Mutex          // mutex protects renderer from concurrent access.
}

func (bro *BrowserPhantomJS) IsConfigured() bool {
	return bro.Renderers != nil && bro.Renderers.BasePortNumber != 0
}

func (bro *BrowserPhantomJS) SelfTest() error {
	if !bro.IsConfigured() {
		return ErrIncompleteConfig
	}
	if err := bro.Renderers.TestPhantomJSExecutable(); err != nil {
		return fmt.Errorf("BrowserPhantomJS.SelfTest: there is an error with PhantomJS executable - %v", err)

	}
	return nil
}

func (bro *BrowserPhantomJS) Initialise() error {
	/*
		While the interactive browser session can have many instances, this browser session may only
		have one instance.
	*/
	bro.mutex = new(sync.Mutex)
	bro.Renderers.MaxInstances = 1
	if err := bro.Renderers.Initialise(); err != nil {
		return fmt.Errorf("BrowserPhantomJS.Initialise: failed to initialise phantomJS lifecycle manager - %v", err)
	}
	return nil
}

func (bro *BrowserPhantomJS) Trigger() Trigger {
	return ".bp"
}

func (bro *BrowserPhantomJS) Execute(cmd Command) (ret *Result) {
	if errResult := cmd.Trim(); errResult != nil {
		return errResult
	}
	// Make sure there is a browser instance
	bro.mutex.Lock()
	bro.mutex.Unlock()
	if bro.renderer != nil {
		// The retrieved instance may be nil if it was killed due to timeout.
		bro.renderer = bro.Renderers.Retrieve(bro.renderer.Index, bro.renderer.Tag)
	}
	// Start a new instance if the previous instance is gone or was never started
	if bro.renderer == nil {
		var err error
		_, bro.renderer, err = bro.Renderers.Acquire()
		if err != nil {
			return &Result{Error: err}
		}
	}
	// Interpret the command
	params := RegexBrowserCommandAndParam.FindStringSubmatch(cmd.Content)
	if params == nil {
		return &Result{Error: ErrBadBrowserParam}
	}
	var output string
	var err error
	switch params[1] {
	case "f":
		// Go forward
		err = bro.renderer.GoForward()
	case "b":
		// Go backward
		err = bro.renderer.GoBack()
	case "p":
		// Go to previous element
		var elements []phantomjs.ElementInfo
		elements, err = bro.renderer.LOPreviousElement()
		output = FormatElementInfoArrayPhantomJS(elements)
	case "n":
		// Go to next element
		var elements []phantomjs.ElementInfo
		elements, err = bro.renderer.LONextElement()
		output = FormatElementInfoArrayPhantomJS(elements)
	case "nn":
		// Go across next N elements
		if len(params) < 3 {
			return &Result{Error: errors.New("usage nn: number")}
		}
		n, err := strconv.Atoi(params[2])
		if err != nil {
			return &Result{Error: errors.New("nn: bad number")}
		}
		var elements []phantomjs.ElementInfo
		elements, err = bro.renderer.LONextNElements(n)
		output = FormatElementInfoArrayPhantomJS(elements)
	case "0":
		// Reset navigation back to the first DOM element
		err = bro.renderer.LOResetNavigation()
	case "r":
		// Reload current page
		err = bro.renderer.Reload()
	case "k":
		// Kill browser
		bro.renderer.Kill()
		bro.renderer = nil
		output = "killed"
	case "g":
		// Go to new URL
		if len(params) < 3 {
			return &Result{Error: errors.New("usage g: url")}
		}
		// Hard code dimension for now, it does not really matter.
		err = bro.renderer.GoTo(phantomjs.GoodUserAgent, params[2], 2560, 1440)
	case "i":
		// Get page info
		var info phantomjs.RemotePageInfo
		info, err = bro.renderer.GetPageInfo()
		output = fmt.Sprintf("%s-%s", info.Title, info.URL)
	case "ptr":
		// Send pointer event to current element
		if len(params) < 3 {
			return &Result{Error: errors.New("usage ptr: type button")}
		}
		typeAndButton := strings.Split(params[2], " ")
		if len(typeAndButton) < 2 {
			return &Result{Error: errors.New("usage ptr: type button")}
		}
		actionType := typeAndButton[0]
		button := typeAndButton[1]
		err = bro.renderer.LOPointer(actionType, button)
	case "val":
		// Give current element a new value
		if len(params) < 3 {
			return &Result{Error: errors.New("usage val: new value")}
		}
		err = bro.renderer.LOSetValue(params[2])
	case "e":
		// Enter text into currently focused element
		if len(params) < 3 {
			return &Result{Error: errors.New("usage e: string")}
		}
		err = bro.renderer.SendKey(params[2], 0)
	case "enter":
		// Press enter key on currently focused element
		err = bro.renderer.SendKey("", phantomjs.KeyCodeEnter)
	case "backsp":
		// Press backspace key on currently focused element
		err = bro.renderer.SendKey("", phantomjs.KeyCodeBackspace)
	case "render":
		// For debugging purpose, render the page screenshot.
		err = bro.renderer.RenderPage()
	default:
		err = ErrBadBrowserParam
	}
	// If there is no other output and no error, result is page info (title - URL).
	if err == nil && output == "" {
		time.Sleep(1 * time.Second)
		var info phantomjs.RemotePageInfo
		info, err = bro.renderer.GetPageInfo()
		output = fmt.Sprintf("%s-%s", info.Title, info.URL)
		if err != nil {
			err = fmt.Errorf("command was successful, but failed to get page info - %v", err)
		}
	}
	return &Result{Error: err, Output: output}
}
