// Time utils to be used from AlpineJS.

// Usage example: <p x-init="renderTSUnixPrettyHour($el)">1717320002</p>
function renderTSUnixPrettyHour(el) {
  renderTSUnixPretty(el, "MMM DD, YYYY, HH:mm")
}

// Usage example: <p x-init="renderTSUnixPrettyNoYear($el)">1717320002</p>
function renderTSUnixPrettyNoYear(el) {
  renderTSUnixPretty(el, "MMM DD, HH:mm")
}

function renderTSUnixPretty(el, format) {
  let ts = dayjs.unix(el.innerHTML)
  el.innerHTML = ts.format(format)
}

// Usage example: <p x-init="renderTSUnixAgo($el)">1717320002</p>
function renderTSUnixAgo(el) {
  let ts = dayjs.unix(el.innerHTML)
  el.innerHTML = ts.fromNow()
}