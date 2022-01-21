package js

import (
	"strconv"
	"time"

	"github.com/go-rod/rod"
)

const GetLinks = `getLinks();
function absolutePath(href) {
    try {
        var link = document.createElement("a");
        link.href = href;
        return link.href.replace('?','/?');
    } catch (error) {}
}
function getLinks() {
    var array = [];
    if (!document) return array;
    var allElements = document.querySelectorAll("*");
    for (var el of allElements) {
        if (el.href && typeof el.href === 'string') {
            array.push(el.href.replace('?','/?'));
        } else if (el.src && typeof el.src === 'string') {
            var absolute = absolutePath(el.src);
            array.push(absolute);
        }
    }
    return array;
}`

func CreateWaitFunc(d time.Duration) *rod.EvalOptions {
	millis := d / time.Millisecond
	return &rod.EvalOptions{
		ByValue:      true,
		JS:           "new Promise(r => setTimeout(r, " + strconv.Itoa(int(millis)) + "));",
		AwaitPromise: true,
	}
}
