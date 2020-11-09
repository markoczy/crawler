package js

const GetLinks = `
function getLinks() {
    var array = [];
    if (!document) return array;
    var allElements = document.querySelectorAll("*");
    for (var el of allElements) {
        if (el.href && typeof el.href === 'string') {
            array.push(el.href);
        } else if (el.src && typeof el.src === 'string') {
            array.push(el.src);
        }
    }
    return array;
}
getLinks();
`
