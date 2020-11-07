package js

const GetLinks = `
function getLinks() {
    var array = [];
    if (!document) return array;
    var allElements = document.querySelectorAll("*");
    for (var el of allElements) {
        if (el.href && typeof el.href === 'string') {
            array.push(el.href);
        }
    }
    return array;
}
getLinks();
`
