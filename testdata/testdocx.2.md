# Markdown Reference

# Markdown For Typora

## Overview

<strong>Markdown</strong> is created by [Daring Fireball](http://daringfireball.net/); the original guideline is [here](http://daringfireball.net/projects/markdown/syntax). Its syntax, however, varies between different parsers or editors. <strong>Typora</strong> is using [GitHub Flavored Markdown](https://help.github.com/articles/github-flavored-markdown/).

[toc]

## Block Elements

### Paragraph and line breaks

A paragraph is simply one or more consecutive lines of text. In markdown source code, paragraphs are separated by two or more blank lines. In Typora, you only need one blank line (press `Return` once) to create a new paragraph.

Press `Shift` + `Return` to create a single line break. Most other markdown parsers will ignore single line breaks, so in order to make other markdown parsers recognize your line break, you can leave two spaces at the end of the line, or insert `<br/>`.

### Headers

Headers use 1-6 hash (`#`) characters at the start of the line, corresponding to header levels 1-6. For example:

```markdown
# This is an H1

## This is an H2

###### This is an H6
```

In Typora, input ‘#’s followed by title content, and press `Return` key will create a header.

### Blockquotes

Markdown uses email-style > characters for block quoting. They are presented as:

```markdown
> This is a blockquote with two paragraphs. This is first paragraph.
>
> This is second pragraph. Vestibulum enim wisi, viverra nec, fringilla in, laoreet vitae, risus.



> This is another blockquote with one paragraph. There is three empty line to seperate two blockquote.
```

In Typora, inputting ‘>’ followed by your quote contents will generate a quote block. Typora will insert a proper ‘>’ or line break for you. Nested block quotes (a block quote inside another block quote) by adding additional levels of ‘>’.

### Lists

Input `* list item 1` will create an unordered list - the `*` symbol can be replace with `+` or `-`.

Input `1. list item 1` will create an ordered list - their markdown source code is as follows:

```markdown
## un-ordered list
*   Red
*   Green
*   Blue

## ordered list
1.  Red
2.         Green
3.        Blue
```

### Task List

Task lists are lists with items marked as either [ ] or [x] (incomplete or complete). For example:

```markdown
- [ ] a task list item
- [ ] list syntax required
- [ ] normal **formatting**, @mentions, #1234 refs
- [ ] incomplete
- [x] completed
```

You can change the complete/incomplete state by clicking on the checkbox before the item.

### (Fenced) Code Blocks

Typora only supports fences in GitHub Flavored Markdown. Original code blocks in markdown are not supported.

Using fences is easy: Input ``and press `return`. Add an optional language identifier after`` and we'll run it through syntax highlighting:

```
Here's an example:

```js
function test() {
  console.log("notice the blank line before this function?");
}
```

syntax highlighting:

```ruby
require 'redcarpet'
markdown = Redcarpet.new("Hello World!")
puts markdown.to_html
```

```
### Math Blocks

You can render <em>LaTeX</em> mathematical expressions using <strong>MathJax</strong>.

To add a mathematical expression, input `$$` and press the 'Return' key. This will trigger an input field which accepts <em>Tex/LaTex</em> source. For example:

$$\mathbf{V}_1 \times \mathbf{V}_2 = \begin{vmatrix}\mathbf{i} & \mathbf{j} & \mathbf{k} \\\frac{\partial X}{\partial u} & \frac{\partial Y}{\partial u} & 0 \\\frac{\partial X}{\partial v} & \frac{\partial Y}{\partial v} & 0 \\\end{vmatrix}$$

In the markdown source file, the math block is a <em>LaTeX</em> expression wrapped by a pair of ‘$$’ marks:

```markdown
$$
\mathbf{V}_1 \times \mathbf{V}_2 =  \begin{vmatrix}
\mathbf{i} & \mathbf{j} & \mathbf{k} \\
\frac{\partial X}{\partial u} &  \frac{\partial Y}{\partial u} & 0 \\
\frac{\partial X}{\partial v} &  \frac{\partial Y}{\partial v} & 0 \\
\end{vmatrix}
$$
```

You can find more details [here](https://support.typora.io/Math/).

### Tables

Input `| First Header | Second Header |` and press the `return` key. This will create a table with two columns.

After a table is created, putting focus on that table will open up a toolbar for the table where you can resize, align, or delete the table. You can also use the context menu to copy and add/delete individual columns/rows.

The full syntax for tables is described below, but it is not necessary to know the full syntax in detail as the markdown source code for tables is generated automatically by Typora.

In markdown source code, they look like:

```markdown
| First Header  | Second Header |
| ------------- | ------------- |
| Content Cell  | Content Cell  |
| Content Cell  | Content Cell  |
```

You can also include inline Markdown such as links, bold, italics, or strikethrough in the table.

Finally, by including colons (`:`) within the header row, you can define text in that column to be left-aligned, right-aligned, or center-aligned:

```markdown
| Left-Aligned  | Center Aligned  | Right Aligned |
| :------------ |:---------------:| -----:|
| col 3 is      | some wordy text | $1600 |
| col 2 is      | centered        |   $12 |
| zebra stripes | are neat        |    $1 |
```

A colon on the left-most side indicates a left-aligned column; a colon on the right-most side indicates a right-aligned column; a colon on both sides indicates a center-aligned column.

### Footnotes

```markdown
You can create footnotes like this[^footnote].

[^footnote]: Here is the *text* of the **footnote**.
```

will produce:

You can create footnotes like this[1].

Hover over the ‘footnote’ superscript to see content of the footnote.

### Horizontal Rules

Inputting `***` or `---` on a blank line and pressing `return` will draw a horizontal line.

### YAML Front Matter

Typora now supports [YAML Front Matter](http://jekyllrb.com/docs/frontmatter/). Input `---` at the top of the article and then press `Return` to introduce a metadata block. Alternatively, you can insert a metadata block from the top menu of Typora.

### Table of Contents (TOC)

Input `[toc]` and press the `Return` key. This will create a “Table of Contents” section. The TOC extracts all headers from the document, and its contents are updated automatically as you add to the document.

## Span Elements

Span elements will be parsed and rendered right after typing. Moving the cursor in middle of those span elements will expand those elements into markdown source. Below is an explanation of the syntax for each span element.

### Links

Markdown supports two styles of links: inline and reference.

In both styles, the link text is delimited by [square brackets].

To create an inline link, use a set of regular parentheses immediately after the link text’s closing square bracket. Inside the parentheses, put the URL where you want the link to point, along with an optional title for the link, surrounded in quotes. For example:

```markdown
This is [an example](http://example.com/ "Title") inline link.

[This link](http://example.net/) has no title attribute.
```

will produce:

This is [an example](http://example.com/) inline link. (`<p>This is <a href="http://example.com/" title="Title">`)

[This link](http://example.net/) has no title attribute. (`<p><a href="http://example.net/">This link</a> has no`)

#### Internal Links

<strong>You can set the href to headers</strong>, which will create a bookmark that allow you to jump to that section after clicking. For example:

Command(on Windows: Ctrl) + Click This link will jump to header `Block Elements`. To see how to write that, please move cursor or click that link with `⌘` key pressed to expand the element into markdown source.

#### Reference Links

Reference-style links use a second set of square brackets, inside which you place a label of your choosing to identify the link:

```markdown
This is [an example][id] reference-style link.

Then, anywhere in the document, you define your link label on a line by itself like this:

[id]: http://example.com/  "Optional Title Here"
```

In Typora, they will be rendered like so:

This is [an example](http://example.com/) reference-style link.

The implicit link name shortcut allows you to omit the name of the link, in which case the link text itself is used as the name. Just use an empty set of square brackets — for example, to link the word “Google” to the google.com web site, you could simply write:

```markdown
[Google][]
And then define the link:

[Google]: http://google.com/
```

In Typora, clicking the link will expand it for editing, and command+click will open the hyperlink in your web browser.

### URLs

Typora allows you to insert URLs as links, wrapped by `<` brackets `>`.

`<i@typora.io>` becomes [i@typora.io](https://mailto:i@typora.io).

Typora will also automatically link standard URLs. e.g: [www.google.com](http://www.google.com).

### Images

Images have similar syntax as links, but they require an additional `!` char before the start of the link. The syntax for inserting an image looks like this:

```markdown
![Alt text](/path/to/img.jpg)

![Alt text](/path/to/img.jpg "Optional title")
```

You are able to use drag & drop to insert an image from an image file or your web browser. You can modify the markdown source code by clicking on the image. A relative path will be used if the image that is added using drag & drop is in same directory or sub-directory as the document you're currently editing.

If you’re using markdown for building websites, you may specify a URL prefix for the image preview on your local computer with property `typora-root-url` in YAML Front Matters. For example, input `typora-root-url:/User/Abner/Website/typora.io/` in YAML Front Matters, and then `![alt](/blog/img/test.png)` will be treated as `![alt](file:///User/Abner/Website/typora.io/blog/img/test.png)` in Typora.

You can find more details [here](https://support.typora.io/Images/).

### Emphasis

Markdown treats asterisks (`*`) and underscores (`_`) as indicators of emphasis. Text wrapped with one `*` or `_` will be wrapped with an HTML `<em>` tag. E.g:

```markdown
*single asterisks*

_single underscores_
```

output:

<em>single asterisks</em>

<em>single underscores</em>

GFM will ignore underscores in words, which is commonly used in code and names, like this:

> wow_great_stuff

> do_this_and_do_that_and_another_thing.

To produce a literal asterisk or underscore at a position where it would otherwise be used as an emphasis delimiter, you can backslash escape it:

```markdown
\*this text is surrounded by literal asterisks\*
```

Typora recommends using the `*` symbol.

### Strong

A double `*` or `_` will cause its enclosed contents to be wrapped with an HTML `<strong>` tag, e.g:

```markdown
**double asterisks**

__double underscores__
```

output:

<strong>double asterisks</strong>

<strong>double underscores</strong>

Typora recommends using the `**` symbol.

### Code

To indicate an inline span of code, wrap it with backtick quotes (`). Unlike a pre-formatted code block, a code span indicates code within a normal paragraph. For example:

```markdown
Use the `printf()` function.
```

will produce:

Use the `printf()` function.

### Strikethrough

GFM adds syntax to create strikethrough text, which is missing from standard Markdown.

`~~Mistaken text.~~` becomes <del>Mistaken text.</del>

### Underlines

Underline is powered by raw HTML.

`<u>Underline</u>` becomes <u>Underline</u>.

### Emoji :smile:

Input emoji with syntax `:smile:`.

User can trigger auto-complete suggestions for emoji by pressing `ESC` key, or trigger it automatically after enabling it on preference panel. Also, inputting UTF-8 emoji characters directly is also supported by going to `Edit` -> `Emoji & Symbols` in the menu bar (macOS).

### Inline Math

To use this feature, please enable it first in the `Preference` Panel -> `Markdown` Tab. Then, use `$` to wrap a TeX command. For example: `$\lim_{x \to \infty} \exp(-x) = 0$` will be rendered as LaTeX command.

To trigger inline preview for inline math: input “$”, then press the `ESC` key, then input a TeX command.

You can find more details [here](https://support.typora.io/Math/).

### Subscript

To use this feature, please enable it first in the `Preference` Panel -> `Markdown` Tab. Then, use `~` to wrap subscript content. For example: `H~2~O`, `X~long\ text~`/

### Superscript

To use this feature, please enable it first in the `Preference` Panel -> `Markdown` Tab. Then, use `^` to wrap superscript content. For example: `X^2^`.

### Highlight

To use this feature, please enable it first in the `Preference` Panel -> `Markdown` Tab. Then, use `==` to wrap highlight content. For example: `==highlight==`.

## HTML

You can use HTML to style content what pure Markdown does not support. For example, use `<span style="color:red">this text is red</span>` to add text with red color.

### Embed Contents

Some websites provide iframe-based embed code which you can also paste into Typora. For example:

```markdown
<iframe height='265' scrolling='no' title='Fancy Animated SVG Menu' src='http://codepen.io/jeangontijo/embed/OxVywj/?height=265&theme-id=0&default-tab=css,result&embed-version=2' frameborder='no' allowtransparency='true' allowfullscreen='true' style='width: 100%;'></iframe>
```

### Video

You can use the `<video>` HTML tag to embed videos. For example:

```markdown
<video src="xxx.mp4" />
```

### Other HTML Support

You can find more details [here](https://support.typora.io/HTML/).

[1]Here is the text of the footnote. 
