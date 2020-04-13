/*
 * CODE GENERATED AUTOMATICALLY WITH
 *    github.com/wlbr/templify
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package reporter

// template_reportTemplate is a generated function returning the template as a string.
// That string should be parsed by the functions of the golang's template package.
func template_reportTemplate() string {
	var tmpl = "{{- /*gotype: github.com/pantheon-systems/search-secrets/pkg/reporter.reportData*/ -}}\n" +
		"{{$devEnabled:=.DevEnabled}}\n" +
		"{{define \"link\"}}<a href=\"{{.URL}}\" title=\"{{.Tooltip}}\" data-toggle=\"tooltip\"\n" +
		"                    data-placement=\"top\">{{.Label}}</a>{{end}}\n" +
		"<!DOCTYPE html>\n" +
		"<html lang=\"en\">\n" +
		"<head>\n" +
		"    <meta charset=\"UTF-8\">\n" +
		"    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">\n" +
		"    <title>Search Secrets Report {{.ReportDate.Format \"01/02/2006 15:04:05\"}}</title>\n" +
		"    <link href=\"https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css\" rel=\"stylesheet\"\n" +
		"          integrity=\"sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh\" crossorigin=\"anonymous\">\n" +
		"    <link href=\"https://fonts.googleapis.com/icon?family=Material+Icons\" rel=\"stylesheet\">\n" +
		"    <style>\n" +
		"        body {\n" +
		"            font-size: 14px;\n" +
		"        }\n" +
		"\n" +
		"        pre {\n" +
		"            padding: 15px;\n" +
		"            background-color: #cccccc;\n" +
		"        }\n" +
		"\n" +
		"        .report-info {\n" +
		"            margin-bottom: 30px;\n" +
		"        }\n" +
		"\n" +
		"        .secret {\n" +
		"            padding-top: 30px;\n" +
		"            margin-bottom: 50px;\n" +
		"            border-top: 3px solid #e9e9e9;\n" +
		"        }\n" +
		"\n" +
		"        .secret > h3 {\n" +
		"            margin-bottom: 30px;\n" +
		"        }\n" +
		"\n" +
		"        .secret-value {\n" +
		"            margin-bottom: 30px;\n" +
		"        }\n" +
		"\n" +
		"        .secret-value > h4 {\n" +
		"            font-size: 14px;\n" +
		"        }\n" +
		"\n" +
		"        .expander-col {\n" +
		"            width: 40px;\n" +
		"        }\n" +
		"\n" +
		"        .expander a:hover {\n" +
		"            text-decoration: none;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full {\n" +
		"            display: table-row;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full.collapsed {\n" +
		"            display: none;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full > th,\n" +
		"        .finding-full > td {\n" +
		"            background-color: #e9e9e9;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full > .expander-col {\n" +
		"            background-color: #fff;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full > td {\n" +
		"            padding: 0\n" +
		"        }\n" +
		"\n" +
		"        .finding-full table tr:first-child th,\n" +
		"        .finding-full table tr:first-child td {\n" +
		"            border-top-width: 0\n" +
		"        }\n" +
		"\n" +
		"        .finding-full table pre {\n" +
		"            margin-bottom: 0\n" +
		"        }\n" +
		"    </style>\n" +
		"</head>\n" +
		"<body>\n" +
		"<div class=\"container-fluid\">\n" +
		"    <h2>Search Secrets Report</h2>\n" +
		"    <table class=\"table report-info\">\n" +
		"        <tr>\n" +
		"            <th scope=\"row\">Secrets found</th>\n" +
		"            <td>{{ len .Secrets}}</td>\n" +
		"        </tr>\n" +
		"        <tr>\n" +
		"            <th scope=\"row\">Completed</th>\n" +
		"            <td>{{.ReportDate.Format \"01/02/2006 15:04:05\"}}</td>\n" +
		"        </tr>\n" +
		"        {{if .Secrets}}\n" +
		"            <tr>\n" +
		"                <th scope=\"row\">Repos with secrets</th>\n" +
		"                <td>\n" +
		"                    {{range $index, $repoName := .Repos}}{{if $index}}, {{end}}{{$repoName}}{{end}}\n" +
		"                </td>\n" +
		"            </tr>\n" +
		"        {{end}}\n" +
		"    </table>\n" +
		"\n" +
		"    {{if not .Secrets}}\n" +
		"        <p>No secrets were found.</p>\n" +
		"    {{end}}\n" +
		"\n" +
		"    {{range $, $secret := .Secrets}}\n" +
		"        <div class=\"secret\">\n" +
		"            <h3>Secret {{$secret.ID}}</h3>\n" +
		"\n" +
		"            <div class=\"secret-value\">\n" +
		"                <h4>Value</h4>\n" +
		"                <pre><code>{{$secret.Value}}</code></pre>\n" +
		"            </div>\n" +
		"\n" +
		"            {{if $secret.ValueDecoded}}\n" +
		"                <div class=\"secret-value secret-value-decoded\">\n" +
		"                    <h4>Decoded</h4>\n" +
		"                    <pre><code>{{$secret.ValueDecoded}}</code></pre>\n" +
		"                </div>\n" +
		"            {{end}}\n" +
		"\n" +
		"            <table class=\"table table-sm findings\">\n" +
		"                <tr>\n" +
		"                    <th scope=\"col\" class=\"expander-col\"></th>\n" +
		"                    <th scope=\"col\">Rule</th>\n" +
		"                    <th scope=\"col\">Repo</th>\n" +
		"                    <th scope=\"col\">Commit</th>\n" +
		"                    <th scope=\"col\">File/Line</th>\n" +
		"                    <th scope=\"col\">Date</th>\n" +
		"                    <th scope=\"col\">Author</th>\n" +
		"                </tr>\n" +
		"                {{range $, $finding := .Findings}}\n" +
		"                    <tr class=\"finding\">\n" +
		"                        <td class=\"expander\">\n" +
		"                            <a href=\"javascript:\" class=\"material-icons\"></a>\n" +
		"                        </td>\n" +
		"                        <td>{{$finding.RuleName}}</td>\n" +
		"                        <td>{{template \"link\" $finding.RepoFullLink}}</td>\n" +
		"                        <td>{{template \"link\" $finding.CommitHashLinkShort}}</td>\n" +
		"                        <td>{{template \"link\" $finding.FileLineLinkShort}}</td>\n" +
		"                        <td>{{$finding.CommitDate.Format \"01/02/2006\"}}</td>\n" +
		"                        <td>{{$finding.CommitAuthorEmail}}</td>\n" +
		"                    </tr>\n" +
		"                    <tr class=\"finding-full collapsed\">\n" +
		"                        <td class=\"expander-col\"></td>\n" +
		"                        <td colspan=\"6\">\n" +
		"                            <table class=\"table table-sm\">\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Rule</th>\n" +
		"                                    <td>{{$finding.RuleName}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Repo</th>\n" +
		"                                    <td>{{template \"link\" $finding.RepoFullLink}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Commit</th>\n" +
		"                                    <td>{{template \"link\" $finding.CommitHashLink}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">File/Line</th>\n" +
		"                                    <td>{{template \"link\" $finding.FileLineLink}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Date</th>\n" +
		"                                    <td>{{$finding.CommitDate.Format \"01/02/2006 15:04:05\"}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Author</th>\n" +
		"                                    <td>{{$finding.CommitAuthorFull}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Code</th>\n" +
		"                                    <td>\n" +
		"                                        <pre><code>{{ $finding.CodeTrimmed }}</code></pre>\n" +
		"                                    </td>\n" +
		"                                </tr>\n" +
		"                                {{if $devEnabled}}\n" +
		"                                    <tr>\n" +
		"                                        <th scope=\"row\">Code</th>\n" +
		"                                        <td>\n" +
		"                                        <pre><code>Repo = \"infrastructure\"\n" +
		"Commit = \"{{$finding.CommitHash}}\"\n" +
		"Path = \"{{$finding.FilePath}}\"\n" +
		"Rule = \"{{$finding.RuleName}}\"\n" +
		"DiffLine = {{$finding.StartLineNumDiff}}</code></pre>\n" +
		"                                        </td>\n" +
		"                                    </tr>\n" +
		"                                {{end}}\n" +
		"\n" +
		"                            </table>\n" +
		"                        </td>\n" +
		"                    </tr>\n" +
		"                {{end}}\n" +
		"            </table>\n" +
		"        </div>\n" +
		"    {{end}}\n" +
		"</div>\n" +
		"<script src=\"https://code.jquery.com/jquery-3.4.1.slim.min.js\"\n" +
		"        integrity=\"sha384-J6qa4849blE2+poT4WnyKhv5vZF5SrPo0iEjwBvKU7imGFAV0wwj1yYfoRSJoZ+n\"\n" +
		"        crossorigin=\"anonymous\"></script>\n" +
		"<script src=\"https://cdn.jsdelivr.net/npm/popper.js@1.16.0/dist/umd/popper.min.js\"\n" +
		"        integrity=\"sha384-Q6E9RHvbIyZFJoft+2mJbHaEWldlvI9IOYy5n3zV9zzTtmI3UksdQRVvoxMfooAo\"\n" +
		"        crossorigin=\"anonymous\"></script>\n" +
		"<script src=\"https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/js/bootstrap.min.js\"\n" +
		"        integrity=\"sha384-wfSDF2E50Y2D1uUdj0O3uMBJnjuUD4Ih7YwaYd1iqfktj0Uod8GCExl3Og8ifwB6\"\n" +
		"        crossorigin=\"anonymous\"></script>\n" +
		"<script type=\"application/javascript\">\n" +
		"    $(function () {\n" +
		"        $('[data-toggle=\"tooltip\"]').tooltip();\n" +
		"\n" +
		"        $('.finding').each(function () {\n" +
		"            const collapsedClass = 'collapsed';\n" +
		"            let finding = $(this);\n" +
		"            let findingFull = $(this).next('.finding-full');\n" +
		"            let expander = finding.find('.expander');\n" +
		"            let expanderLink = expander.children('a');\n" +
		"\n" +
		"            function updateIcon() {\n" +
		"                expanderLink[0].innerHTML = findingFull.hasClass(collapsedClass) ? 'add_circle' : 'remove_circle';\n" +
		"            }\n" +
		"\n" +
		"            function toggleCollapsed() {\n" +
		"                findingFull.toggleClass(collapsedClass);\n" +
		"                updateIcon()\n" +
		"            }\n" +
		"\n" +
		"            expanderLink.click(toggleCollapsed);\n" +
		"            updateIcon()\n" +
		"        })\n" +
		"    });\n" +
		"</script>\n" +
		"</body>\n" +
		"</html>\n" +
		""
	return tmpl
}
