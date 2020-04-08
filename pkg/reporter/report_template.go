/*
 * CODE GENERATED AUTOMATICALLY WITH
 *    github.com/wlbr/templify
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package reporter

// report_templateTemplate is a generated function returning the template as a string.
// That string should be parsed by the functions of the golang's template package.
func report_templateTemplate() string {
	var tmpl = "{{- /*gotype: github.com/pantheon-systems/search-secrets/pkg/reporter.reportData*/ -}}\n" +
		"{{define \"link\"}}<a href=\"{{.URL}}\" title=\"{{.Tooltip}}\" data-toggle=\"tooltip\"\n" +
		"                    data-placement=\"top\">{{.Label}}</a>{{end}}\n" +
		"<!DOCTYPE html>\n" +
		"<html lang=\"en\">\n" +
		"<head>\n" +
		"    <meta charset=\"UTF-8\">\n" +
		"    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">\n" +
		"    <title>Title</title>\n" +
		"    <link href=\"https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css\" rel=\"stylesheet\"\n" +
		"          integrity=\"sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh\" crossorigin=\"anonymous\">\n" +
		"    <link href=\"https://fonts.googleapis.com/icon?family=Material+Icons\" rel=\"stylesheet\">\n" +
		"    <style>\n" +
		"        body {\n" +
		"            font-size: 14px;\n" +
		"        }\n" +
		"\n" +
		"        pre {\n" +
		"            padding: 5px;\n" +
		"            background-color: gainsboro;\n" +
		"        }\n" +
		"\n" +
		"        .finding-code pre {\n" +
		"            max-width: 200px;\n" +
		"            max-height: 200px;\n" +
		"            text-overflow: ellipsis;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full {\n" +
		"            display: table-row;\n" +
		"        }\n" +
		"\n" +
		"        .finding-full.collapsed {\n" +
		"            display: none;\n" +
		"        }\n" +
		"    </style>\n" +
		"</head>\n" +
		"<body>\n" +
		"<div class=\"container-fluid\">\n" +
		"    {{range .Secrets}}\n" +
		"        <div class=\"secret\" style=\"margin-bottom: 50px\">\n" +
		"            <h3>Secret {{.ID}}</h3>\n" +
		"\n" +
		"            <div class=\"secret-value\">\n" +
		"                <pre><code>{{.Value}}</code></pre>\n" +
		"            </div>\n" +
		"\n" +
		"            <h4>Findings</h4>\n" +
		"            <table class=\"table findings\">\n" +
		"                <tr>\n" +
		"                    <th scope=\"col\"></th>\n" +
		"                    <th scope=\"col\">Rule</th>\n" +
		"                    <th scope=\"col\">Repo</th>\n" +
		"                    <th scope=\"col\">Commit</th>\n" +
		"                    <th scope=\"col\">Date</th>\n" +
		"                    <th scope=\"col\">Author</th>\n" +
		"                    <th scope=\"col\">Line</th>\n" +
		"                </tr>\n" +
		"                {{range $i, $finding := .Findings}}\n" +
		"                    <tr class=\"finding\">\n" +
		"                        <td class=\"expander\">\n" +
		"                            <a href=\"javascript:\" class=\"material-icons\"></a>\n" +
		"                        </td>\n" +
		"                        <td>{{$finding.RuleName}}</td>\n" +
		"                        <td>{{template \"link\" $finding.RepoFullLink}}</td>\n" +
		"                        <td>{{template \"link\" $finding.CommitHashLinkShort}}</td>\n" +
		"                        <td>{{$finding.CommitDate.Format \"01/02/2006\"}}</td>\n" +
		"                        <td>{{$finding.CommitAuthorEmail}}</td>\n" +
		"                        <td>{{template \"link\" $finding.FileLineLinkShort}}</td>\n" +
		"                    </tr>\n" +
		"                    <tr class=\"finding-full\">\n" +
		"                        <td></td>\n" +
		"                        <td colspan=\"6\">\n" +
		"\n" +
		"                            <table class=\"table finding-full\">\n" +
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
		"                                    <th scope=\"row\">Date</th>\n" +
		"                                    <td>{{$finding.CommitDate.Format \"01/02/2006 15:04:05\"}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Author</th>\n" +
		"                                    <td>{{$finding.CommitAuthorFull}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Line</th>\n" +
		"                                    <td>{{template \"link\" $finding.FileLineLink}}</td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Code</th>\n" +
		"                                    <td>\n" +
		"                                        <pre><code>{{$finding.Code}}</code></pre>\n" +
		"                                    </td>\n" +
		"                                </tr>\n" +
		"                                <tr>\n" +
		"                                    <th scope=\"row\">Diff</th>\n" +
		"                                    <td>\n" +
		"                                        <pre><code>{{$finding.Diff}}</code></pre>\n" +
		"                                    </td>\n" +
		"                                </tr>\n" +
		"                            </table>\n" +
		"\n" +
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
		"            function toggleCollapsed() {\n" +
		"                findingFull.toggleClass(collapsedClass);\n" +
		"                expanderLink[0].innerHTML = findingFull.hasClass(collapsedClass) ? 'add_circle' : 'remove_circle';\n" +
		"            }\n" +
		"\n" +
		"            expanderLink.click(toggleCollapsed);\n" +
		"            toggleCollapsed()\n" +
		"        })\n" +
		"    });\n" +
		"</script>\n" +
		"</body>\n" +
		"</html>\n" +
		""
	return tmpl
}
