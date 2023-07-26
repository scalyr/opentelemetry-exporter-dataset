// Simple utility script which utilizes invert-markdown-table module to invert / transpose
// mmarkdown table which is passed in via stdin.

const imt = require("invert-markdown-table")

let data = "";

async function main() {
  for await (const chunk of process.stdin) data += chunk;
  console.log(imt(data));
}

main();
