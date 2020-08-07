package statusreporter

// This file defines the static web pages and the text of the HTML templates.

// Define the static web pages
var stylesheetPage = []byte(`
div.notice {
  color: green;
  font-weight: bold;
}

div.ErrorMessage {
  color: red;
  font-weight: bold;
}

div.preformatted {
    font-family: monospace;
	white-space: pre;
	display: block;
}
`)

var internalErrorPage = []byte(`
<html>
<head>
	<title>Internal error</title>
</head>
<body>
    <p>
        <font color='red'><b>Internal Error - please try again later</b></font>
    </p>
</body>
</html>
`)

// Define the text for the page templates.
var baseText = `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>{{ template "PageTitle" .}}</title>
        <link href='/{{ template "ServiceName" .}}/stylesheet.css' rel='stylesheet'/>
    </head>
    <body>
    	 <h2>{{ template "PageTitle" .}}</h2>
        <section id="content">
            {{template "content" .}}
        </section>
    </body>
</html>
`

var reportText = `
{{define "PageTitle"}}{{.PageTitle}}{{end}}
{{define "ServiceName"}}{{.ServiceName}}{{end}}
{{define "content"}}
{{.Content}}
{{end}}
`
var errorText = `
{{define "PageTitle"}}Error{{end}}
{{define "content"}}
<h2>Error</h2>
<div class='ErrorMessage' id='ErrorMessage'>
	<b>{{.}}</b>
</div>
{{end}}
`
