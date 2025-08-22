# Vanilla Framework Syntax
Vanilla 是一个全栈式 Web 框架，前端 UI 使用 HTML、CSS 和 JavaScript 构建，后端逻辑则采用 Go 实现。作为一个典型的服务端渲染（SSR，Server-Side Rendering）框架，Vanilla 的 HTML 内容均由后端生成，类似于传统的 JSP、PHP 或 Django 架构。

Vanilla 也是一个基于编译期的框架。其核心工具 —— Vanilla Compiler 会在构建过程中将 HTML 组件模板编译为 Go 源码，最终生成一个单一的可执行文件。该文件通过 Go 的 embed 技术，内嵌所有所需的静态资源（如 CSS、JS、图片等），从而实现无需依赖外部资源的部署体验。

## Project Structure
Here’s how a Vanilla project is organized.

### Directories and Files

Vanilla leverages an opinionated folder layout for your project. Every Vanilla project root should include the following directories and files:

```
bin/* - Your project artifacts.
pages/* - Your project source code (components, pages, styles, images, etc.)
public/* - Your non-code, unprocessed assets (fonts, icons, etc.)
package.json - A project manifest.
vanilla.options.json - A Vanilla configuration file. (optional)
```

### Project Tree
A common Vanilla project directory might look like this:
```
bin/
public/
    robots.txt
    favicon.svg
pages/
    Layout.html
    Index.html
    router.go
    blog/
        Index.html
        route.go
        stlye.css
package.json
vanilla.options.json
```


## UI
Vanilla 使用 HTML、CSS 和 Javascript 来构建 UI，不同于传统的是，Vanilla 引入了组件系统。
每个 Vanilla 组件都可以看成是一个 HTML 文件，Vanilla 只是对 HTML 语法进行了增强，并设定了一些限制。

### 组件文件
Vanilla 组件沿用了标准 HTML 类型文件的扩展名`.html`。具体设定如下：

1. 只有放置在项目顶层 `pages/` 目录下的 HTML 文件才会被当成 Vanilla 组件处理。否则 Vanilla Compiler & Vanilla IDE Extension 会忽略这些文件。
2. 所有组件文件名必须采用大驼峰命名（如 `Button.HTML`、`SideBar.HTML`）。Vanilla Compiler & Vanilla IDE Extension 基于这个规则来区别一个标签是 HTML 标签还是组件标签。

### 组件内容布局
Vanilla leverages an opinionated content layout for components.

1. 每个组件必须且只能包含一个顶层 `Javascript 代码块` 和一个顶层 `Template 代码块`。
2. `Javascript 代码块` 代码块由一个顶层`<script>`标签包裹。`Template 代码块`由一个块级 HTML 标签（比如`<div></div>`）包裹。如果你需要一个空的`Template 代码块`包裹标签，可以使用`<fragment></fragment>`.
3. `Javascript 代码块` 必须为于组件内容最顶部，且内容必须是 ES Module 语法，Vanilla Compiler & Vanilla IDE Extension 默认这个代码块中的JS使用了严格模式。
4. 除了`Javascript 代码块`和 `Template 代码块`外，组件内不能再有其他`顶层`标签，也不能有`顶层`注释。
5. 除了顶层的`Javascript 代码块`，组件内部不允许使用其他内联`Javascript`代码的`<script>`标签。但是支持`<script src="my.js"></script>`这种语法。

有效的组件内容布局如下：
`Example01.html`
```HTML
<script>
    // ES Module code block
</script>

<div>
    <!--Template code block-->
    <span>
        Hello World.
    </span>
    <script src="external.js"></script>
</div>
```

`Example02.html`
```HTML
<script>
    // ES Module code block
</script>

<head>
    <!--Template code block-->
    <style>
        body{
            background: black;
        }
    </style>
</head>
```

`Example03.html`
```HTML
<script>
    // ES Module code block
</script>

<fragment>
    <!--Template code block-->
    <title>My First Vanilla Page</title>
    <meta charset="UTF-8"/>
</fragment>
```

### 组件导入
Vanilla 支持在`Javascript 代码块`中使用 `import` 语句引入其他组件。Vanilla 支持两种导入方式：

副作用导入（bare import）：如 `import "./Item.html"`。

默认导入（default import）：如 `import Card from "./Card.html"`，主要用于组件名冲突时。

示例如下：
```HTML
<script>
    import "./Item.html"
    import Card from "./Card.html"
</script>

<div>
    <Card>
        <Item>Hello World</Item>
    </Card>
</div>
```

### 组件属性
在 Vanilla 中，组件属性是一种特殊的模板变量，在这个设定中，每个属性都来源于某个类型，要么来自 Go 的类型，要么来自 JS 的字面量类型。
Vanilla 也称组件属性为`Macro Variable`，因为这些声明这些组件属性的`Macro`会在编译和运行时展开。

#### 属性声明
Vanilla 在`Javascript 代码块`置入了一个关键字`extern`，extern 看起来就像一个JS函数（不过它是一个`Macro`），组件属性通过这个宏来声明。

Vanilla 组件的属性声明必须符合以下规范：
1. `extern macro`接受参数只能是 Go 的composite types(Struct、Map、Slice)的实例，或JS字面量类型实例，比如boolean, number, string, empty array(aka []), empty object(aka {})。
2. 属性声明必须使用let关键字来。
3. 属性声明不支持JS的解构赋值语法。

示例如下：
```HTML
<script>
    import {User} from "./user.go"
    // Equals Go's User type instance `User{}`.
    let user = extern(User())
    // Equals Go's []any type instance `make([]any,0)`.
    let tags = extern([])
    // Equals Go's map[string]any type instance `make(map[string]any)`.
    let jsonObject = extern({})
    // Equals Go's string type instance `dark`.
    let theme = extern("dark")
    // Equals Go's bool type instance `false`.
    let disable = extern(false)
    // Equals Go's int32 type instance `1`.
    let counts = extern(1)
    // Equals Go's float32 type instance `0.8`.
    let discount = extern(0.8)
</script>

<div>
    <Card>
        <Item>{user.name}</Item>
    </Card>
</div>
```

#### 属性访问
Vanilla 支持在`Template 代码块`中使用标准的属性访问（如 `{user.disabled}`）和索引访问（如 `{user.tags[0]}`）方式去访问属性变量的成员。
但不支持 Hash 访问语法（如 `{user["myKey"]}`）。

Vanilla 同样支持一个简单的 boolean unary 表达式访问，比如`{!user.active}`。

```HTML
<script>
    import {User} from "./user.go"
    let user = extern(User())
</script>

<div>
    <Card>
        <Item>name: {user.name}</Item>
    </Card>
    <button disabled={!user.active}></button>
</div>
```

#### 属性赋值
在 Vanilla 中，如果一个组件是一个顶层组件，那么它的属性赋值一定来自模版渲染器（在 Go 的运行时环境中）。
其他情况下，组件属性，可以通过类似 HTML Attribute 的方式进行赋值。
或通过一个内置`<context>`组件来跨越层级赋值。

`Top.html`
```HTML
<script>
    let theme = extern("dark")
</script>

<div>
    <context theme={theme}>
        <Parent></Parent>
    </context>
</div>
```

`Parent.html`
```HTML
<script>
    import {User} from "./user.go"
    let user = extern(User())
</script>

<div>
    <Children tags={user.tags}></Children>
</div>
```

`Children.html`
```HTML
<script>
    import {Tags} from "./tag.go"
    // Accepts the value from `Parent.html` component.
    let tags = extern(Tags())
    // Will be overwritten by `Top.html` component.
    let theme = extern("light")
</script>

<div class="{theme}">
    <div>
        User Tags: {for _, tag in tags}<span>{tag}</span>{/for}
    </div>
</div>
```

### 模板函数
Vanilla 组件支持在模板中使用几个必须的函数。函数不支持嵌套。

### 流程控制
#### If 语句
支持 if 条件判断，语句内仅允许逻辑表达式与比较表达式，例如：
{if !user.disabled && user.likes > 0}
支持 else 分支，但不支持 else if。

#### For 语句
支持 for 循环，可遍历集合或数值区间，如：
{for index, value in user.tags} 或 {for i, v in 1..9}

```HTML
<script>
    import {User} from "./user.go"
    import Card from "./Card.HTML"
    let user = extern(User())
</script>

<div class="{user.theme}">
    {if !user.disabled}
        <Card>
            <div>name:<span>{user.name}</span></div>
            <div>age:<span>{user.age}</span></div>
            <div>tag:
                {for index, value in user.tags}
                <span data-index={index}>{value}</span>
                {/for}
            </div>
        </Card>
    {else}
        <Card>User[{user.name}] disabled</Card>
    {/if}
</div>
```

### 样式表
Vanilla 支持通过 import 语句导入 CSS 样式表。 如 `import "./style.css"`。 不过需要注意，组件导入 CSS 样式表是在全局上下文中内有效的。

示例如下：
```HTML
<script>
    import "./style.css"
</script>

<div>
    <Card>
        Hello World.
    </Card>
</div>
```


## Routing
Vanilla 提供固执己见的路由功能。

### Router
### Route


















