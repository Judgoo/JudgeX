wrk.method = "POST"
wrk.body = [[
{
	"id": 1,
	"code": "#include <iostream>\nusing namespace std;\n\nint main() {\n    cout << \"helloworld\" << endl;\n    return 0;\n}\n",
	"inputs": [
		"",
		"123",
		"2"
	],
	"outputs": [
		"helloworld\n",
		"helloworld1\n",
		"helloworld\n"
	]
}
]]
wrk.headers["Content-Type"] = "application/json"
