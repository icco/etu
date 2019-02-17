import MarkdownIt from "markdown-it";

let md = new MarkdownIt({
  // Enable HTML tags in source
  html: true,

  // Use '/' to close single tags (<br />).
  xhtmlOut: true,

  // Convert '\n' in paragraphs into <br>
  breaks: true,

  // Autoconvert URL-like text to links
  linkify: true,
}).use(require("markdown-it-footnote"));

// Add Twitter User support
md.linkify.add("@", {
  validate: function(text, pos, self) {
    var tail = text.slice(pos);

    if (!self.re.twitter) {
      self.re.twitter = new RegExp(
        "^([a-zA-Z0-9_]){1,15}(?!_)(?=$|" + self.re.src_ZPCc + ")"
      );
    }
    if (self.re.twitter.test(tail)) {
      // Linkifier allows punctuation chars before prefix,
      // but we additionally disable `@` ("@@mention" is invalid)
      if (pos >= 2 && tail[pos - 2] === "@") {
        return false;
      }
      return tail.match(self.re.twitter)[0].length;
    }
    return 0;
  },
  normalize: function(match) {
    match.url = "https://twitter.com/" + match.url.replace(/^@/, "");
  },
});

// Add hashtag support
md.linkify.add("#", {
  validate: function(text, pos, self) {
    var tail = text.slice(pos);

    if (!self.re.twitter) {
      self.re.twitter = new RegExp(
        "^([a-zA-Z0-9_]){1,15}(?!_)(?=$|" + self.re.src_ZPCc + ")"
      );
    }
    if (self.re.twitter.test(tail)) {
      // Linkifier allows punctuation chars before prefix,
      // but we additionally disable `#` ("##tag" is invalid)
      if (pos >= 2 && tail[pos - 2] === "@") {
        return false;
      }
      return tail.match(self.re.twitter)[0].length;
    }
    return 0;
  },
  normalize: function(match) {
    match.url = "/tag/" + match.url.replace(/^#/, "");
  },
});

export default md;
