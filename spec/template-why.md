# Template Engine

String token just works in if statement !!!

No raw string supported in If Statement, why 👇

// {if var1 == "abc"}
// Some case like this {if var1 == "<a"} conflicts with the HTML start tag open token.
// We can use {if var1 == "&lt;abc"} instead to avoid ambiguity.
// But {if var1 == `abc`}, we cannot do any thing.


No string supported in html tag attribute, why 👇

```html
<div title="{user["name"]}"></div>
```

{var1.a.b} -> 👌
{var1.a[3].b} -> 👌
{var1["name"]} -> 🙅
{var1['name']} -> 🙅



