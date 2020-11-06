function getLinks() {
    var array = [];
    var allElements = document.querySelectorAll("*");
    for (var el of allElements) {
        if (el.href && typeof el.href === 'string') {
            array.push(el.href);
        }
    }
    return array;
}
getLinks();