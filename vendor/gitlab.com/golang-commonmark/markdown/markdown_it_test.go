// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"fmt"
	"testing"
)

var mditTests = []struct {
	markdown string
	want     string
	filename string
}{
	// 0
	// Issue #246.  Double escaping in ALT
	{"![&](#)\n", "<p><img src=\"#\" alt=\"&amp;\"></p>\n", "commonmark_extras.txt"},

	// 1
	// Strip markdown in ALT tags
	{"\n![*strip* [markdown __in__ alt](#)](#)\n", "<p><img src=\"#\" alt=\"strip markdown in alt\"></p>\n", "commonmark_extras.txt"},

	// 2
	// Issue #55:
	{"\n![test]\n\n![test](foo bar)\n", "<p>![test]</p>\n<p>![test](foo bar)</p>\n", "commonmark_extras.txt"},

	// 3
	// Issue #35. `<` should work as punctuation
	{"\nan **(:**<br>\n", "<p>an <strong>(:</strong><br></p>\n", "commonmark_extras.txt"},

	// 4
	// Should unescape only needed things in link destinations/titles:
	{"\n[test](<\\f\\o\\o\\>\\\\>)\n\n[test](foo \"\\\\\\\"\\b\\a\\r\")\n", "<p><a href=\"%5Cf%5Co%5Co%3E%5C\">test</a></p>\n<p><a href=\"foo\" title=\"\\&quot;\\b\\a\\r\">test</a></p>\n", "commonmark_extras.txt"},

	// 5
	// Not a closing tag
	{"\n</ 123>\n", "<p>&lt;/ 123&gt;</p>\n", "commonmark_extras.txt"},

	// 6
	// Escaping entities in links:
	{"\n[](<&quot;> \"&amp;&ouml;\")\n\n[](<\\&quot;> \"\\&amp;\\&ouml;\")\n\n[](<\\\\&quot;> \"\\\\&quot;\\\\&ouml;\")\n", "<p><a href=\"%22\" title=\"&amp;ö\"></a></p>\n<p><a href=\"&amp;quot;\" title=\"&amp;amp;&amp;ouml;\"></a></p>\n<p><a href=\"%5C%22\" title=\"\\&quot;\\ö\"></a></p>\n", "commonmark_extras.txt"},

	// 7
	// Checking combination of replaceEntities and unescapeMd:
	{"\n~~~ &amp;&bad;\\&amp;\\\\&amp;\njust a funny little fence\n~~~\n", "<pre><code class=\"&amp;&amp;bad;&amp;amp;\\&amp;\">just a funny little fence\n</code></pre>\n", "commonmark_extras.txt"},

	// 8
	// Underscore between punctuation chars should be able to close emphasis.
	{"\n_(hai)_.\n", "<p><em>(hai)</em>.</p>\n", "commonmark_extras.txt"},

	// 9
	// Those are two separate blockquotes:
	{"\n - > foo\n  > bar\n", "<ul>\n<li>\n<blockquote>\n<p>foo</p>\n</blockquote>\n</li>\n</ul>\n<blockquote>\n<p>bar</p>\n</blockquote>\n", "commonmark_extras.txt"},

	// 10
	// Blockquote should terminate itself after paragraph continuation
	{"\n- list\n    > blockquote\nblockquote continuation\n    - next list item\n", "<ul>\n<li>list\n<blockquote>\n<p>blockquote\nblockquote continuation</p>\n</blockquote>\n<ul>\n<li>next list item</li>\n</ul>\n</li>\n</ul>\n", "commonmark_extras.txt"},

	// 11
	// Regression test (code block + regular paragraph)
	{"\n>     foo\n> bar\n", "<blockquote>\n<pre><code>foo\n</code></pre>\n<p>bar</p>\n</blockquote>\n", "commonmark_extras.txt"},

	// 12
	// Blockquotes inside indented lists should terminate correctly
	{"\n   - a\n     > b\n     ```\n     c\n     ```\n   - d\n", "<ul>\n<li>a\n<blockquote>\n<p>b</p>\n</blockquote>\n<pre><code>c\n</code></pre>\n</li>\n<li>d</li>\n</ul>\n", "commonmark_extras.txt"},

	// 13
	// Don't output empty class here:
	{"\n```&#x20;\ntest\n```\n", "<pre><code>test\n</code></pre>\n", "commonmark_extras.txt"},

	// 14
	// Setext header text supports lazy continuations:
	{"\n - foo\nbar\n   ===\n", "<ul>\n<li>\n<h1>foo\nbar</h1>\n</li>\n</ul>\n", "commonmark_extras.txt"},

	// 15
	// But setext header underline doesn't:
	{"\n - foo\n   bar\n  ===\n", "<ul>\n<li>foo\nbar\n===</li>\n</ul>\n", "commonmark_extras.txt"},

	// 16
	// Info string in fenced code block can't contain marker used for the fence
	{"\n~~~test~\n\n~~~test`\n", "<p>~~~test~</p>\n<pre><code class=\"test`\"></code></pre>\n", "commonmark_extras.txt"},

	// 17
	// Tabs should be stripped from the beginning of the line
	{"\n foo\n bar\n\tbaz\n", "<p>foo\nbar\nbaz</p>\n", "commonmark_extras.txt"},

	// 18
	// Tabs should not cause hardbreak, EOL tabs aren't stripped in commonmark 0.27
	{"\nfoo1\t\nfoo2    \nbar\n", "<p>foo1\t\nfoo2<br>\nbar</p>\n", "commonmark_extras.txt"},

	// 19
	// List item terminating quote should not be paragraph continuation
	{"\n1. foo\n   > quote\n2. bar\n", "<ol>\n<li>foo\n<blockquote>\n<p>quote</p>\n</blockquote>\n</li>\n<li>bar</li>\n</ol>\n", "commonmark_extras.txt"},

	// 20
	// Coverage. Directive can terminate paragraph.
	{"\na\n<?php\n", "<p>a</p>\n<?php\n", "commonmark_extras.txt"},

	// 21
	// Coverage. Nested email autolink (silent mode)
	{"\n*<foo@bar.com>*\n", "<p><em><a href=\"mailto:foo@bar.com\">foo@bar.com</a></em></p>\n", "commonmark_extras.txt"},

	// 22
	// Coverage. Unpaired nested backtick (silent mode)
	{"\n*`foo*\n", "<p><em>`foo</em></p>\n", "commonmark_extras.txt"},

	// 23
	// Coverage. Entities.
	{"\n*&*\n\n*&#x20;*\n\n*&amp;*\n", "<p><em>&amp;</em></p>\n<p><em> </em></p>\n<p><em>&amp;</em></p>\n", "commonmark_extras.txt"},

	// 24
	// Coverage. Escape.
	{"\n*\\a*\n", "<p><em>\\a</em></p>\n", "commonmark_extras.txt"},

	// 25
	// Coverage. parseLinkDestination
	{"\n[foo](<\nbar>)\n\n[foo](<bar)\n", "<p>[foo](&lt;\nbar&gt;)</p>\n<p>[foo](&lt;bar)</p>\n", "commonmark_extras.txt"},

	// 26
	// Coverage. parseLinkTitle
	{"\n[foo](bar \"ba)\n\n[foo](bar \"ba\\\nz\")\n", "<p>[foo](bar &quot;ba)</p>\n<p><a href=\"bar\" title=\"ba\\\nz\">foo</a></p>\n", "commonmark_extras.txt"},

	// 27
	// Coverage. Image
	{"\n![test]( x )\n", "<p><img src=\"x\" alt=\"test\"></p>\n", "commonmark_extras.txt"},

	// 28
	// Coverage. Image
	{"![test][foo]\n\n[bar]: 123\n", "<p>![test][foo]</p>\n", "commonmark_extras.txt"},

	// 29
	// Coverage. Image
	{"![test][[[\n\n[bar]: 123\n", "<p>![test][[[</p>\n", "commonmark_extras.txt"},

	// 30
	// Coverage. Image
	{"![test](\n", "<p>![test](</p>\n", "commonmark_extras.txt"},

	// 31
	// Coverage. Link
	{"\n[test](\n", "<p>[test](</p>\n", "commonmark_extras.txt"},

	// 32
	// Coverage. Reference
	{"\n[\ntest\\\n]: 123\nfoo\nbar\n", "<p>foo\nbar</p>\n", "commonmark_extras.txt"},

	// 33
	// Coverage. Reference
	{"[\ntest\n]\n", "<p>[\ntest\n]</p>\n", "commonmark_extras.txt"},

	// 34
	// Coverage. Reference
	{"> [foo]: bar\n[foo]\n", "<blockquote>\n</blockquote>\n<p><a href=\"bar\">foo</a></p>\n", "commonmark_extras.txt"},

	// 35
	// Coverage. Tabs in blockquotes.
	{"\n>\t\ttest\n\n >\t\ttest\n\n  >\t\ttest\n\n> ---\n>\t\ttest\n\n > ---\n >\t\ttest\n\n  > ---\n  >\t\ttest\n\n>\t\t\ttest\n\n >\t\t\ttest\n\n  >\t\t\ttest\n\n> ---\n>\t\t\ttest\n\n > ---\n >\t\t\ttest\n\n  > ---\n  >\t\t\ttest\n", "<blockquote>\n<pre><code>  test\n</code></pre>\n</blockquote>\n<blockquote>\n<pre><code> test\n</code></pre>\n</blockquote>\n<blockquote>\n<pre><code>test\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code>  test\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code> test\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code>test\n</code></pre>\n</blockquote>\n<blockquote>\n<pre><code>  \ttest\n</code></pre>\n</blockquote>\n<blockquote>\n<pre><code> \ttest\n</code></pre>\n</blockquote>\n<blockquote>\n<pre><code>\ttest\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code>  \ttest\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code> \ttest\n</code></pre>\n</blockquote>\n<blockquote>\n<hr>\n<pre><code>\ttest\n</code></pre>\n</blockquote>\n", "commonmark_extras.txt"},

	// 36
	// Coverage. Tabs in lists.
	{"\n1. \tfoo\n\n\t     bar\n", "<ol>\n<li>\n<p>foo</p>\n<pre><code> bar\n</code></pre>\n</li>\n</ol>\n", "commonmark_extras.txt"},

	// 37
	// Coverage. Various tags not interrupting blockquotes because of indentation:
	{"\n> foo\n    - - - -\n\n> foo\n    # not a heading\n\n> foo\n    ```\n    not a fence\n    ```\n", "<blockquote>\n<p>foo\n- - - -</p>\n</blockquote>\n<blockquote>\n<p>foo\n# not a heading</p>\n</blockquote>\n<blockquote>\n<p>foo\n<code>not a fence</code></p>\n</blockquote>\n", "commonmark_extras.txt"},

	// 38
	// Should not throw exception on invalid chars in URL (`*` not allowed in path) [mailformed URI]
	{"[foo](<&#x25;test>)\n", "<p><a href=\"%25test\">foo</a></p>\n", "fatal.txt"},

	// 39
	// Should not throw exception on broken utf-8 sequence in URL [mailformed URI]
	{"\n[foo](%C3)\n", "<p><a href=\"%C3\">foo</a></p>\n", "fatal.txt"},

	// 40
	// Should not throw exception on broken utf-16 surrogates sequence in URL [mailformed URI]
	{"\n[foo](&#xD800;)\n", "<p><a href=\"&amp;#xD800;\">foo</a></p>\n", "fatal.txt"},

	// 41
	// Should not hang comments regexp
	{"\nfoo <!--- xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx ->\n\nfoo <!------------------------------------------------------------------->\n", "<p>foo &lt;!— xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -&gt;</p>\n<p>foo &lt;!-------------------------------------------------------------------&gt;</p>\n", "fatal.txt"},

	// 42
	// Should not hang cdata regexp
	{"\nfoo <![CDATA[ xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx ]>\n", "<p>foo &lt;![CDATA[ xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx ]&gt;</p>\n", "fatal.txt"},

	// 43
	// linkify
	{"url http://www.youtube.com/watch?v=5Jt5GEr4AYg.\n", "<p>url <a href=\"http://www.youtube.com/watch?v=5Jt5GEr4AYg\">http://www.youtube.com/watch?v=5Jt5GEr4AYg</a>.</p>\n", "linkify.txt"},

	// 44
	// don't touch text in links
	{"\n[https://example.com](https://example.com)\n", "<p><a href=\"https://example.com\">https://example.com</a></p>\n", "linkify.txt"},

	// 45
	// don't touch text in autolinks
	{"\n<https://example.com>\n", "<p><a href=\"https://example.com\">https://example.com</a></p>\n", "linkify.txt"},

	// 46
	// don't touch text in html <a> tags
	{"\n<a href=\"https://example.com\">https://example.com</a>\n", "<p><a href=\"https://example.com\">https://example.com</a></p>\n", "linkify.txt"},

	// 47
	// match links without protocol
	{"\nwww.example.org\n", "<p><a href=\"http://www.example.org\">www.example.org</a></p>\n", "linkify.txt"},

	// 48
	// emails
	{"\ntest@example.com\n\nmailto:test@example.com\n", "<p><a href=\"mailto:test@example.com\">test@example.com</a></p>\n<p><a href=\"mailto:test@example.com\">mailto:test@example.com</a></p>\n", "linkify.txt"},

	// 49
	// typorgapher should not break href
	{"\nhttp://example.com/(c)\n", "<p><a href=\"http://example.com/(c)\">http://example.com/(c)</a></p>\n", "linkify.txt"},

	// 50
	// Encode link destination, decode text inside it:
	{"<http://example.com/α%CE%B2γ%CE%B4>\n", "<p><a href=\"http://example.com/%CE%B1%CE%B2%CE%B3%CE%B4\">http://example.com/αβγδ</a></p>\n", "normalize.txt"},

	// 51
	{"\n[foo](http://example.com/α%CE%B2γ%CE%B4)\n", "<p><a href=\"http://example.com/%CE%B1%CE%B2%CE%B3%CE%B4\">foo</a></p>\n", "normalize.txt"},

	// 52
	// Should decode punycode:
	{"\n<http://xn--n3h.net/>\n", "<p><a href=\"http://xn--n3h.net/\">http://☃.net/</a></p>\n", "normalize.txt"},

	// 53
	{"\n<http://☃.net/>\n", "<p><a href=\"http://xn--n3h.net/\">http://☃.net/</a></p>\n", "normalize.txt"},

	// 54
	// Invalid punycode:
	{"\n<http://xn--xn.com/>\n", "<p><a href=\"http://xn--xn.com/\">http://xn--xn.com/</a></p>\n", "normalize.txt"},

	// 55
	// Invalid punycode (non-ascii):
	{"\n<http://xn--γ.com/>\n", "<p><a href=\"http://xn--xn---emd.com/\">http://xn--γ.com/</a></p>\n", "normalize.txt"},

	// 56
	// Two slashes should start a domain:
	{"\n[](//☃.net/)\n", "<p><a href=\"//xn--n3h.net/\"></a></p>\n", "normalize.txt"},

	// 57
	// Don't encode domains in unknown schemas:
	{"\n[](skype:γγγ)\n", "<p><a href=\"skype:%CE%B3%CE%B3%CE%B3\"></a></p>\n", "normalize.txt"},

	// 58
	// Should auto-add protocol to autolinks:
	{"\ntest google.com foo\n", "<p>test <a href=\"http://google.com\">google.com</a> foo</p>\n", "normalize.txt"},

	// 59
	// Should support IDN in autolinks:
	{"\ntest http://xn--n3h.net/ foo\n", "<p>test <a href=\"http://xn--n3h.net/\">http://☃.net/</a> foo</p>\n", "normalize.txt"},

	// 60
	{"\ntest http://☃.net/ foo\n", "<p>test <a href=\"http://xn--n3h.net/\">http://☃.net/</a> foo</p>\n", "normalize.txt"},

	// 61
	{"\ntest //xn--n3h.net/ foo\n", "<p>test <a href=\"//xn--n3h.net/\">//☃.net/</a> foo</p>\n", "normalize.txt"},

	// 62
	{"\ntest xn--n3h.net foo\n", "<p>test <a href=\"http://xn--n3h.net\">☃.net</a> foo</p>\n", "normalize.txt"},

	// 63
	{"\ntest xn--n3h@xn--n3h.net foo\n", "<p>test <a href=\"mailto:xn--n3h@xn--n3h.net\">xn--n3h@☃.net</a> foo</p>\n", "normalize.txt"},

	// 64
	{"[__proto__]\n\n[__proto__]: blah\n", "<p><a href=\"blah\"><strong>proto</strong></a></p>\n", "proto.txt"},

	// 65
	{"\n[hasOwnProperty]\n\n[hasOwnProperty]: blah\n", "<p><a href=\"blah\">hasOwnProperty</a></p>\n", "proto.txt"},

	// 66
	// Should parse nested quotes:
	{"\"foo 'bar' baz\"\n\n'foo 'bar' baz'\n", "<p>“foo ‘bar’ baz”</p>\n<p>‘foo ‘bar’ baz’</p>\n", "smartquotes.txt"},

	// 67
	// Should not overlap quotes:
	{"\n'foo \"bar' baz\"\n", "<p>‘foo &quot;bar’ baz&quot;</p>\n", "smartquotes.txt"},

	// 68
	// Should match quotes on the same level:
	{"\n\"foo *bar* baz\"\n", "<p>“foo <em>bar</em> baz”</p>\n", "smartquotes.txt"},

	// 69
	// Should handle adjacent nested quotes:
	{"\n'\"double in single\"'\n\n\"'single in double'\"\n", "<p>‘“double in single”’</p>\n<p>“‘single in double’”</p>\n", "smartquotes.txt"},

	// 70
	// Should not match quotes on different levels:
	{"\n*\"foo* bar\"\n\n\"foo *bar\"*\n\n*\"foo* bar *baz\"*\n", "<p><em>&quot;foo</em> bar&quot;</p>\n<p>&quot;foo <em>bar&quot;</em></p>\n<p><em>&quot;foo</em> bar <em>baz&quot;</em></p>\n", "smartquotes.txt"},

	// 71
	// Smartquotes should not overlap with other tags:
	{"\n*foo \"bar* *baz\" quux*\n", "<p><em>foo &quot;bar</em> <em>baz&quot; quux</em></p>\n", "smartquotes.txt"},

	// 72
	// Should try and find matching quote in this case:
	{"\n\"foo \"bar 'baz\"\n", "<p>&quot;foo “bar 'baz”</p>\n", "smartquotes.txt"},

	// 73
	// Should not touch 'inches' in quotes:
	{"\n\"Monitor 21\"\" and \"Monitor\"\"\n", "<p>“Monitor 21&quot;” and “Monitor”&quot;</p>\n", "smartquotes.txt"},

	// 74
	// Should render an apostrophe as a rsquo:
	{"\nThis isn't and can't be the best approach to implement this...\n", "<p>This isn’t and can’t be the best approach to implement this…</p>\n", "smartquotes.txt"},

	// 75
	// Apostrophe could end the word, that's why original smartypants replaces all of them as rsquo:
	{"\nusers' stuff\n", "<p>users’ stuff</p>\n", "smartquotes.txt"},

	// 76
	// Quotes between punctuation chars:
	{"\n\"(hai)\".\n", "<p>“(hai)”.</p>\n", "smartquotes.txt"},

	// 77
	// Quotes at the start/end of the tokens:
	{"\n\"*foo* bar\"\n\n\"foo *bar*\"\n\n\"*foo bar*\"\n", "<p>“<em>foo</em> bar”</p>\n<p>“foo <em>bar</em>”</p>\n<p>“<em>foo bar</em>”</p>\n", "smartquotes.txt"},

	// 78
	// Should treat softbreak as a space:
	{"\n\"this\"\nand \"that\".\n\n\"this\" and\n\"that\".\n", "<p>“this”\nand “that”.</p>\n<p>“this” and\n“that”.</p>\n", "smartquotes.txt"},

	// 79
	// Should treat hardbreak as a space:
	{"\n\"this\"\\\nand \"that\".\n\n\"this\" and\\\n\"that\".\n", "<p>“this”<br>\nand “that”.</p>\n<p>“this” and<br>\n“that”.</p>\n", "smartquotes.txt"},

	// 80
	{"~~Strikeout~~\n", "<p><s>Strikeout</s></p>\n", "strikethrough.txt"},

	// 81
	{"\nx ~~~~foo~~ bar~~\n", "<p>x <s><s>foo</s> bar</s></p>\n", "strikethrough.txt"},

	// 82
	{"\nx ~~foo ~~bar~~~~\n", "<p>x <s>foo <s>bar</s></s></p>\n", "strikethrough.txt"},

	// 83
	{"\nx ~~~~foo~~~~\n", "<p>x <s><s>foo</s></s></p>\n", "strikethrough.txt"},

	// 84
	{"\nx ~~a ~~foo~~~~~~~~~~~bar~~ b~~\n\nx ~~a ~~foo~~~~~~~~~~~~bar~~ b~~\n", "<p>x <s>a <s>foo</s></s>~~~<s><s>bar</s> b</s></p>\n<p>x <s>a <s>foo</s></s>~~~~<s><s>bar</s> b</s></p>\n", "strikethrough.txt"},

	// 85
	// Strikeouts have the same priority as emphases:
	{"\n**~~test**~~\n\n~~**test~~**\n", "<p><strong>~~test</strong>~~</p>\n<p><s>**test</s>**</p>\n", "strikethrough.txt"},

	// 86
	// Strikeouts have the same priority as emphases with respect to links:
	{"\n[~~link]()~~\n\n~~[link~~]()\n", "<p><a href=\"\">~~link</a>~~</p>\n<p>~~<a href=\"\">link~~</a></p>\n", "strikethrough.txt"},

	// 87
	// Strikeouts have the same priority as emphases with respect to backticks:
	{"\n~~`code~~`\n\n`~~code`~~\n", "<p>~~<code>code~~</code></p>\n<p><code>~~code</code>~~</p>\n", "strikethrough.txt"},

	// 88
	// Nested strikeouts:
	{"\n~~foo ~~bar~~ baz~~\n\n~~f **o ~~o b~~ a** r~~\n", "<p><s>foo <s>bar</s> baz</s></p>\n<p><s>f <strong>o <s>o b</s> a</strong> r</s></p>\n", "strikethrough.txt"},

	// 89
	// Should not have a whitespace between text and "~~":
	{"\nfoo ~~ bar ~~ baz\n", "<p>foo ~~ bar ~~ baz</p>\n", "strikethrough.txt"},

	// 90
	// Newline should be considered a whitespace:
	{"\n~~test\n~~\n\n~~\ntest~~\n\n~~\ntest\n~~\n", "<p>~~test\n~~</p>\n<p>~~\ntest~~</p>\n<p>~~\ntest\n~~</p>\n", "strikethrough.txt"},

	// 91
	// From CommonMark test suite, replacing `**` with our marker:
	{"\na~~\"foo\"~~\n", "<p>a~~“foo”~~</p>\n", "strikethrough.txt"},

	// 92
	// Simple:
	{"| Heading 1 | Heading 2\n| --------- | ---------\n| Cell 1    | Cell 2\n| Cell 3    | Cell 4\n", "<table>\n<thead>\n<tr>\n<th>Heading 1</th>\n<th>Heading 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell 1</td>\n<td>Cell 2</td>\n</tr>\n<tr>\n<td>Cell 3</td>\n<td>Cell 4</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 93
	// Column alignment:
	{"\n| Header 1 | Header 2 | Header 3 | Header 4 |\n| :------: | -------: | :------- | -------- |\n| Cell 1   | Cell 2   | Cell 3   | Cell 4   |\n| Cell 5   | Cell 6   | Cell 7   | Cell 8   |\n", "<table>\n<thead>\n<tr>\n<th style=\"text-align:center\">Header 1</th>\n<th style=\"text-align:right\">Header 2</th>\n<th style=\"text-align:left\">Header 3</th>\n<th>Header 4</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td style=\"text-align:center\">Cell 1</td>\n<td style=\"text-align:right\">Cell 2</td>\n<td style=\"text-align:left\">Cell 3</td>\n<td>Cell 4</td>\n</tr>\n<tr>\n<td style=\"text-align:center\">Cell 5</td>\n<td style=\"text-align:right\">Cell 6</td>\n<td style=\"text-align:left\">Cell 7</td>\n<td>Cell 8</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 94
	// Nested emphases:
	{"\nHeader 1|Header 2|Header 3|Header 4\n:-------|:------:|-------:|--------\nCell 1  |Cell 2  |Cell 3  |Cell 4\n*Cell 5*|Cell 6  |Cell 7  |Cell 8\n", "<table>\n<thead>\n<tr>\n<th style=\"text-align:left\">Header 1</th>\n<th style=\"text-align:center\">Header 2</th>\n<th style=\"text-align:right\">Header 3</th>\n<th>Header 4</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td style=\"text-align:left\">Cell 1</td>\n<td style=\"text-align:center\">Cell 2</td>\n<td style=\"text-align:right\">Cell 3</td>\n<td>Cell 4</td>\n</tr>\n<tr>\n<td style=\"text-align:left\"><em>Cell 5</em></td>\n<td style=\"text-align:center\">Cell 6</td>\n<td style=\"text-align:right\">Cell 7</td>\n<td>Cell 8</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 95
	// Nested tables inside blockquotes:
	{"\n> foo|foo\n> ---|---\n> bar|bar\nbaz|baz\n", "<blockquote>\n<table>\n<thead>\n<tr>\n<th>foo</th>\n<th>foo</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>bar</td>\n<td>bar</td>\n</tr>\n</tbody>\n</table>\n</blockquote>\n<p>baz|baz</p>\n", "tables.txt"},

	// 96
	// Minimal one-column:
	{"\n| foo\n|----\n| test2\n", "<table>\n<thead>\n<tr>\n<th>foo</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>test2</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 97
	// This is parsed as one big table:
	{"\n-   foo|foo\n---|---\nbar|bar\n", "<table>\n<thead>\n<tr>\n<th>-   foo</th>\n<th>foo</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>bar</td>\n<td>bar</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 98
	// Second line should not contain symbols except "-", ":", "|" and " ":
	{"\nfoo|foo\n---|---s\nbar|bar\n", "<p>foo|foo\n—|—s\nbar|bar</p>\n", "tables.txt"},

	// 99
	// Second line should contain "|" symbol:
	{"\nfoo|foo\n---:---\nbar|bar\n", "<p>foo|foo\n—:—\nbar|bar</p>\n", "tables.txt"},

	// 100
	// Second line should not have empty columns in the middle:
	{"\nfoo|foo\n---||---\nbar|bar\n", "<p>foo|foo\n—||—\nbar|bar</p>\n", "tables.txt"},

	// 101
	// Wrong alignment symbol position:
	{"\nfoo|foo\n---|-::-\nbar|bar\n", "<p>foo|foo\n—|-::-\nbar|bar</p>\n", "tables.txt"},

	// 102
	// Title line should contain "|" symbol:
	{"\nfoo\n---|---\nbar|bar\n", "<p>foo\n—|—\nbar|bar</p>\n", "tables.txt"},

	// 103
	// Allow tabs as a separator on 2nd line
	{"\n|\tfoo\t|\tbar\t|\n|\t---\t|\t---\t|\n|\tbaz\t|\tquux\t|\n", "<table>\n<thead>\n<tr>\n<th>foo</th>\n<th>bar</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>baz</td>\n<td>quux</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 104
	// Should terminate paragraph:
	{"\nparagraph\nfoo|foo\n---|---\nbar|bar\n", "<p>paragraph</p>\n<table>\n<thead>\n<tr>\n<th>foo</th>\n<th>foo</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>bar</td>\n<td>bar</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 105
	// Should be terminated via row without "|" symbol:
	{"\nfoo|foo\n---|---\nparagraph\n", "<table>\n<thead>\n<tr>\n<th>foo</th>\n<th>foo</th>\n</tr>\n</thead>\n<tbody></tbody>\n</table>\n<p>paragraph</p>\n", "tables.txt"},

	// 106
	// Delimiter escaping:
	{"\n| Heading 1 \\\\\\\\| Heading 2\n| --------- | ---------\n| Cell\\|1\\|| Cell\\|2\n\\| Cell\\\\\\|3 \\\\| Cell\\|4\n", "<table>\n<thead>\n<tr>\n<th>Heading 1 \\\\</th>\n<th>Heading 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell|1|</td>\n<td>Cell|2</td>\n</tr>\n<tr>\n<td>| Cell\\|3 \\</td>\n<td>Cell|4</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 107
	// Pipes inside backticks don't split cells:
	{"\n| Heading 1 | Heading 2\n| --------- | ---------\n| Cell 1 | Cell 2\n| `Cell|3` | Cell 4\n", "<table>\n<thead>\n<tr>\n<th>Heading 1</th>\n<th>Heading 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell 1</td>\n<td>Cell 2</td>\n</tr>\n<tr>\n<td><code>Cell|3</code></td>\n<td>Cell 4</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 108
	// Unclosed backticks don't count
	{"\n| Heading 1 | Heading 2\n| --------- | ---------\n| Cell 1 | Cell 2\n| `Cell 3| Cell 4\n", "<table>\n<thead>\n<tr>\n<th>Heading 1</th>\n<th>Heading 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell 1</td>\n<td>Cell 2</td>\n</tr>\n<tr>\n<td>`Cell 3</td>\n<td>Cell 4</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 109
	// Another complicated backticks case
	{"\n| Heading 1 | Heading 2\n| --------- | ---------\n| Cell 1 | Cell 2\n| \\\\\\`|\\\\\\`\n", "<table>\n<thead>\n<tr>\n<th>Heading 1</th>\n<th>Heading 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell 1</td>\n<td>Cell 2</td>\n</tr>\n<tr>\n<td>\\`</td>\n<td>\\`</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 110
	// `\` in tables should not count as escaped backtick
	{"\n# | 1 | 2\n--|--|--\nx | `\\` | `x`\n", "<table>\n<thead>\n<tr>\n<th>#</th>\n<th>1</th>\n<th>2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>x</td>\n<td><code>\\</code></td>\n<td><code>x</code></td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 111
	// Tables should handle escaped backticks
	{"\n# | 1 | 2\n--|--|--\nx | \\`\\` | `x`\n", "<table>\n<thead>\n<tr>\n<th>#</th>\n<th>1</th>\n<th>2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>x</td>\n<td>``</td>\n<td><code>x</code></td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 112
	// An amount of rows might be different across the table (issue #171):
	{"\n| 1 | 2 |\n| :-----: |  :-----: |  :-----: |\n| 3 | 4 | 5 | 6 |\n", "<table>\n<thead>\n<tr>\n<th style=\"text-align:center\">1</th>\n<th style=\"text-align:center\">2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td style=\"text-align:center\">3</td>\n<td style=\"text-align:center\">4</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 113
	// An amount of rows might be different across the table #2:
	{"\n| 1 | 2 | 3 | 4 |\n| :-----: |  :-----: |  :-----: |  :-----: |\n| 5 | 6 |\n", "<table>\n<thead>\n<tr>\n<th style=\"text-align:center\">1</th>\n<th style=\"text-align:center\">2</th>\n<th style=\"text-align:center\">3</th>\n<th style=\"text-align:center\">4</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td style=\"text-align:center\">5</td>\n<td style=\"text-align:center\">6</td>\n<td style=\"text-align:center\"></td>\n<td style=\"text-align:center\"></td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 114
	// Allow one-column tables (issue #171):
	{"\n| foo |\n:-----:\n| bar |\n", "<table>\n<thead>\n<tr>\n<th style=\"text-align:center\">foo</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td style=\"text-align:center\">bar</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 115
	// Allow indented tables (issue #325):
	{"\n  | Col1a | Col2a |\n  | ----- | ----- |\n  | Col1b | Col2b |\n", "<table>\n<thead>\n<tr>\n<th>Col1a</th>\n<th>Col2a</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Col1b</td>\n<td>Col2b</td>\n</tr>\n</tbody>\n</table>\n", "tables.txt"},

	// 116
	// Tables should not be indented more than 4 spaces (1st line):
	{"\n    | Col1a | Col2a |\n  | ----- | ----- |\n  | Col1b | Col2b |\n", "<pre><code>| Col1a | Col2a |\n</code></pre>\n<p>| ----- | ----- |\n| Col1b | Col2b |</p>\n", "tables.txt"},

	// 117
	// Tables should not be indented more than 4 spaces (2nd line):
	{"\n  | Col1a | Col2a |\n    | ----- | ----- |\n  | Col1b | Col2b |\n", "<p>| Col1a | Col2a |\n| ----- | ----- |\n| Col1b | Col2b |</p>\n", "tables.txt"},

	// 118
	// Tables should not be indented more than 4 spaces (3rd line):
	{"\n  | Col1a | Col2a |\n  | ----- | ----- |\n    | Col1b | Col2b |\n", "<table>\n<thead>\n<tr>\n<th>Col1a</th>\n<th>Col2a</th>\n</tr>\n</thead>\n<tbody></tbody>\n</table>\n<pre><code>| Col1b | Col2b |\n</code></pre>\n", "tables.txt"},

	// 119
	// Allow tables with empty body:
	{"\n  | Col1a | Col2a |\n  | ----- | ----- |\n", "<table>\n<thead>\n<tr>\n<th>Col1a</th>\n<th>Col2a</th>\n</tr>\n</thead>\n<tbody></tbody>\n</table>\n", "tables.txt"},

	// 120
	// Align row should be at least as large as any actual rows:
	{"\nCol1a | Col1b | Col1c\n----- | -----\nCol2a | Col2b | Col2c\n", "<p>Col1a | Col1b | Col1c\n----- | -----\nCol2a | Col2b | Col2c</p>\n", "tables.txt"},

	// 121
	{"(bad)\n", "<p>(bad)</p>\n", "typographer.txt"},

	// 122
	// copyright
	{"\n(c) (C)\n", "<p>© ©</p>\n", "typographer.txt"},

	// 123
	// reserved
	{"\n(r) (R)\n", "<p>® ®</p>\n", "typographer.txt"},

	// 124
	// trademark
	{"\n(tm) (TM)\n", "<p>™ ™</p>\n", "typographer.txt"},

	// 125
	// paragraph
	{"\n(p) (P)\n", "<p>§ §</p>\n", "typographer.txt"},

	// 126
	// plus-minus
	{"\n+-5\n", "<p>±5</p>\n", "typographer.txt"},

	// 127
	// ellipsis
	{"\ntest.. test... test..... test?..... test!....\n", "<p>test… test… test… test?.. test!..</p>\n", "typographer.txt"},

	// 128
	// dupes
	{"\n!!!!!! ???? ,,\n", "<p>!!! ??? ,</p>\n", "typographer.txt"},

	// 129
	// dashes
	{"\n---markdownit --- super---\n\nmarkdownit---awesome\n\nabc ----\n\n--markdownit -- super--\n\nmarkdownit--awesome\n", "<p>—markdownit — super—</p>\n<p>markdownit—awesome</p>\n<p>abc ----</p>\n<p>–markdownit – super–</p>\n<p>markdownit–awesome</p>\n", "typographer.txt"},

	// 130
	{"[normal link](javascript)\n", "<p><a href=\"javascript\">normal link</a></p>\n", "xss.txt"},

	// 131
	// Should not allow some protocols in links and images
	{"\n[xss link](javascript:alert(1))\n\n[xss link](JAVASCRIPT:alert(1))\n\n[xss link](vbscript:alert(1))\n\n[xss link](VBSCRIPT:alert(1))\n\n[xss link](file:///123)\n", "<p>[xss link](javascript:alert(1))</p>\n<p>[xss link](JAVASCRIPT:alert(1))</p>\n<p>[xss link](vbscript:alert(1))</p>\n<p>[xss link](VBSCRIPT:alert(1))</p>\n<p>[xss link](file:///123)</p>\n", "xss.txt"},

	// 132
	{"\n[xss link](&#34;&#62;&#60;script&#62;alert&#40;&#34;xss&#34;&#41;&#60;/script&#62;)\n\n[xss link](&#74;avascript:alert(1))\n\n[xss link](&#x26;#74;avascript:alert(1))\n\n[xss link](\\&#74;avascript:alert(1))\n", "<p><a href=\"%22%3E%3Cscript%3Ealert(%22xss%22)%3C/script%3E\">xss link</a></p>\n<p>[xss link](Javascript:alert(1))</p>\n<p><a href=\"&amp;#74;avascript:alert(1)\">xss link</a></p>\n<p><a href=\"&amp;#74;avascript:alert(1)\">xss link</a></p>\n", "xss.txt"},

	// 133
	{"\n[xss link](<javascript:alert(1)>)\n", "<p>[xss link](&lt;javascript:alert(1)&gt;)</p>\n", "xss.txt"},

	// 134
	{"\n[xss link](javascript&#x3A;alert(1))\n", "<p>[xss link](javascript:alert(1))</p>\n", "xss.txt"},

	// 135
	// Should not allow data-uri except some whitelisted mimes
	{"\n![](data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7)\n", "<p><img src=\"data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7\" alt=\"\"></p>\n", "xss.txt"},

	// 136
	{"\n[xss link](data:text/html;base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K)\n", "<p>[xss link](data:text/html;base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K)</p>\n", "xss.txt"},

	// 137
	{"\n[normal link](/javascript:link)\n", "<p><a href=\"/javascript:link\">normal link</a></p>\n", "xss.txt"},

	// 138
	// Image parser use the same code base as link.
	{"\n![xss link](javascript:alert(1))\n", "<p>![xss link](javascript:alert(1))</p>\n", "xss.txt"},

	// 139
	// Autolinks
	{"\n<javascript&#x3A;alert(1)>\n\n<javascript:alert(1)>\n", "<p>&lt;javascript:alert(1)&gt;</p>\n<p>&lt;javascript:alert(1)&gt;</p>\n", "xss.txt"},

	// 140
	// Linkifier
	{"\njavascript&#x3A;alert(1)\n\njavascript:alert(1)\n", "<p>javascript:alert(1)</p>\n<p>javascript:alert(1)</p>\n", "xss.txt"},

	// 141
	// References
	{"\n[test]: javascript:alert(1)\n", "<p>[test]: javascript:alert(1)</p>\n", "xss.txt"},

	// 142
	// Make sure we decode entities before split:
	{"\n```js&#32;custom-class\ntest1\n```\n\n```js&#x0C;custom-class\ntest2\n```\n", "<pre><code class=\"js\">test1\n</code></pre>\n<pre><code class=\"js\">test2\n</code></pre>\n", "xss.txt"},
}

func TestMarkdownIt(t *testing.T) {
	for i, test := range mditTests {
		switch i {
		case 35: // XXX opennota: investigate
			continue
		}

		test := test
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			t.Parallel()
			md := New(HTML(true), LangPrefix(""))
			got := md.RenderToString([]byte(test.markdown))
			if got != test.want {
				t.Errorf("#%d: markdown(%q)\nwant %q\n got %q", i, test.markdown, test.want, got)
			}
		})
	}
}
