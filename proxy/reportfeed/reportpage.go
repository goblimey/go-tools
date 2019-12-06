package reportfeed

// reportFormat defines the HTML structure of the report.
const reportFormat = `
<h3>Last Client Buffer</h3>
<span id='clienttimestamp'>%s</span>
<pre>
<code>
<div class="preformatted" id='clientbuffer'>
%s
</div>
</code>
</pre>
<pre>
<code>
<h3>Last Server Buffer</h3>
<span id='servertimestamp'>%s</span>
<div class="preformatted" id='serverbuffer'>
%s
</div>
</code>
</pre>
`