package friendscript

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PerformLine/friendscript/utils"
	"github.com/PerformLine/go-stockutil/httputil"
	"github.com/PerformLine/go-stockutil/maputil"
	"github.com/PerformLine/go-stockutil/typeutil"
	"github.com/stretchr/testify/require"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

type testCommands struct {
	utils.Module
	env utils.Runtime
}

func newTestCommands(env utils.Runtime) *testCommands {
	cmd := &testCommands{
		env: env,
	}

	cmd.Module = utils.NewDefaultExecutor(cmd)
	return cmd
}

func (self *testCommands) MapArg(key string, m map[string]interface{}) error {
	self.env.Scope().Set(strings.TrimSpace(key), m)
	return nil
}

func (self *testCommands) Noop() error {
	return nil
}

func eval(script string, items ...interface{}) (map[string]interface{}, error) {
	env := NewEnvironment()
	env.RegisterModule(`testing`, newTestCommands(env))

	if len(items) > 0 {
		script = fmt.Sprintf(script, items...)
	}

	scope, err := env.EvaluateString(script)

	if err == nil {
		return scope.Data(), nil
	} else {
		return nil, err
	}
}

func TestListCommands(t *testing.T) {
	assert := require.New(t)
	env := NewEnvironment()
	env.RegisterModule(`testing`, newTestCommands(env))
	commands := env.Commands()

	assert.Contains(commands, `core::log`)
	assert.Contains(commands, `file::read`)
	assert.Contains(commands, `fmt::trim`)
	assert.Contains(commands, `testing::map_arg`)
	assert.Contains(commands, `testing::noop`)
	assert.Contains(commands, `vars::get`)
}

func TestAssignments(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`null`: nil,
		`a`:    1,
		`b`:    true,
		`c`:    "Test",
		`c1`:   "Test Test",
		`c2`:   "Test {c}",
		`d`:    3.14159,
		`e`: []interface{}{
			float64(1),
			true,
			"Test",
			3.14159,
			[]interface{}{
				float64(1),
				true,
				"Test",
				3.14159,
			},
			map[string]interface{}{
				`ok`: true,
			},
		},
		`f`: map[string]interface{}{
			`ok`: true,
		},
		`g`: `g`,
		`h`: `h`,
		`i`: `i`,
		`j`: `j`,
		`k`: nil,
		`l`: `l`,
		`m`: `m`,
		`o`: `o`,
		`p`: nil,
		`q`: `q`,
		`s`: `s`,
		`t`: true,
		`u`: nil,
		`v`: 1,
		`w`: true,
		`x`: "Test",
		`y`: 3.14159,
		`z`: []interface{}{
			float64(1),
			true,
			"Test",
			3.14159,
		},
		`z0`:     1,
		`z1`:     true,
		`z2`:     `Test`,
		`z3`:     3.14159,
		`aa`:     1,
		`bb`:     true,
		`cc`:     "Test",
		`dd`:     3.14159,
		`eeType`: `pi-3.14159`,
		`ee0`: map[string]interface{}{
			`ok`: true,
			`always`: map[string]interface{}{
				`finishing`: map[string]interface{}{
					`each_others`: `sentences`,
				},
			},
			`Content-Type`: `delicatessen/bologna`,
			`pi-3.14159`:   `yep`,
		},
		`ee`: map[string]interface{}{
			`ok`: true,
			`always`: map[string]interface{}{
				`finishing`: map[string]interface{}{
					`each_others`: `sentences`,
				},
			},
		},
		`ee1`: map[string]interface{}{
			`finishing`: map[string]interface{}{
				`each_others`: `sentences`,
			},
		},
		`ee2`: map[string]interface{}{
			`each_others`: `sentences`,
		},
		`ee3`:  `sentences`,
		`ee4`:  nil,
		`ee5`:  true,
		`ee6`:  `sentences`,
		`ekey`: `always`,
		`ee7`:  `sentences`,
		`ee8`: map[string]interface{}{
			`ok`: true,
			`always`: map[string]interface{}{
				`finishing`: map[string]interface{}{
					`each_others`: `sandwiches`,
					`other`: map[string]interface{}{
						`stuff`: map[string]interface{}{
							`too`: true,
						},
					},
				},
			},
		},
		`put_1`: `test 1`,
		`put_2`: `test {a}`,
		`put_3`: []interface{}{
			float64(1),
			float64(2),
			float64(3),
		},
		`put_4`: "put test four\n\t\tput test\n\t\tput end\n\t\tend friend end",
		`t_maparg`: map[string]interface{}{
			`one`:   `first`,
			`two`:   `second`,
			`three`: `third`,
		},
		`ulit1`: `\u2211`,
		`ulit2`: "\u2211",
		`vars_set_1`: []interface{}{
			`set1`,
		},
		`stringify_float`: `2.718281828459045`,
		`stringify_int`:   `8675309`,
	}

	script := `# set variables of with values of every type
    $null = null
    $a = 1
    $b = true
    $c = "Test"
    $c1 = "Test {c}"
    $c2 = 'Test {c}'
    $d = 3.14159
    $e = [1, true, "Test", 3.14159, [1, true, "Test", 3.14159], {ok:true}]
    $f = {
        ok: true,
    }
    $g, $h, $i = "g", "h", "i"
    $j, $k = "j"
    $l, $m = ["l", "m", "n"]
    $o, $p = ["o"]
    $q, _, $s = ["q", "r", "s"]
    $t = $f.ok
    $u = $f.nonexistent
    $v, $w, $x, $y, $z = [1, true, "Test", 3.14159, [1, true, "Test", 3.14159]]
    $z0 = $z[0]
    $z1 = $z[1]
    $z2 = $z[2]
    $z3 = $z[3]
    # capture command results as variables, and also put a bunch of them on the same line
    put 1 -> $aa; put true -> $bb; put "Test" -> $cc; put 3.14159 -> $dd
	$eeType = "pi-{d}"
    $ee0 = {
        'ok': true,
        "always": {
            'finishing': {
                "each_others": "sentences",
            },
        },
		"Content-Type": "delicatessen/bologna",
		"{eeType}": "yep",
    }

    put {
        ok: true,
        always: {
            finishing: {
                each_others: "sentences",
            },
        },
    } -> $ee
    $ee1, $ee2 = $ee.always, $ee.always.finishing
    $ee3, $ee4 = [$ee.always.finishing.each_others, $ee.always.finishing.each_others.sandwiches]
    $ee5 = $ee['ok']
    $ee6 = $ee['always'].finishing['each_others']
    $ekey = 'always'
    $ee7 = $ee[$ekey].finishing['each_others']
    $ee8 = $ee
    $ee8.always['finishing'].each_others = 'sandwiches'
    $ee8.always['finishing'].other['stuff'].too = true
    put "test {a}" -> $put_1
    put 'test {a}' -> $put_2
	put [1, 2, 3] -> $put_3
	put """
		put test four
		put test
		put end
		end friend end

	""" -> $put_4
	testing::map_arg 't_maparg' {
		one:   'first',
		two:   'second',
		three: 'third',
	}
	$ulit1 = '\u2211'
	$ulit2 = "\u2211"

	vars::set 'vars_set_1' {
		value: ['set1'],
	}

	$stringify_float = 2.718281828459045
	$stringify_float = "{stringify_float}"
	$stringify_int = 8675309
	$stringify_int = "{stringify_int}"`

	actual, err := eval(script)

	assert.NoError(err)

	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestIfScopes(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`a`:             `top_a`,
		`b`:             `top_b`,
		`a_if`:          `top_a`,
		`b_if`:          `if_b`,
		`a_if_if`:       `if_if_a`,
		`b_if_if`:       `if_b`,
		`a_after_if_if`: `top_a`,
		`b_after_if_if`: `if_b`,
		`a_after_if`:    `top_a`,
		`b_after_if`:    `top_b`,
		`enter_if_val`:  51,
		`enter_el_val`:  61,
	}

	script := `
        $a             = "top_a"
        $b             = "top_b"
        $a_if          = null
        $b_if          = null
        $a_if_if       = null
        $b_if_if       = null
        $a_after_if_if = null
        $b_after_if_if = null
        $a_after_if    = null
        $b_after_if    = null

        if $b = "if_b"; $b {
            $a_if = $a
            $b_if = $b
            if $a = "if_if_a"; $a {
                $a_if_if = $a
                $b_if_if = $b
            }
            $a_after_if_if = $a
            $b_after_if_if = $b
        }
        $a_after_if = $a
        $b_after_if = $b
        $enter_if_val = null
        $enter_el_val = null

        # if condition trigger, verify condition value, populate via assignment
        if $value = 51; $value > 50 {
            $enter_if_val = 51
        } else {
            $enter_if_val = 9999
        }

        # else condition trigger, verify condition value, populate via command output
        if put 61 -> $value; $value > 100 {
            $enter_el_val = 7777
        } else {
            $enter_el_val = 61
        }`

	actual, err := eval(script)

	assert.NoError(err)

	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestConditionals(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`ten`:            10,
		`unset`:          nil,
		`true`:           true,
		`false`:          false,
		`string`:         "string",
		`names`:          []interface{}{"Bob", "Steve", "Fred"},
		`if_eq`:          true,
		`if_ne`:          true,
		`if_eq_null`:     true,
		`if_true`:        true,
		`if_false`:       true,
		`if_gt`:          true,
		`if_gte`:         true,
		`if_lt`:          true,
		`if_lte`:         true,
		`if_in`:          true,
		`if_not_in`:      true,
		`if_match_1`:     true,
		`if_match_2`:     true,
		`if_match_3`:     true,
		`if_not_match_1`: true,
		`if_not_match_2`: true,
		`if_not_match_3`: true,
		`if_match_4`:     true,
		`if_match_5`:     true,
		`if_match_6`:     true,
		`if_not_match_4`: true,
		`if_not_match_5`: true,
		`if_not_match_6`: true,
	}

	script := `
        $ten = 10
        $unset = null
        $true = true
        $false = false
        $string = "string"
        $names = ["Bob", "Steve", "Fred"]
        $if_eq = null
        $if_ne = null
        $if_eq_null = null
        $if_true = null
        $if_gt = null
        $if_gte = null
        $if_lt = null
        $if_lte = null
        $if_in = null
        $if_not_in = null
        $if_match_1 = null
        $if_match_2 = null
        $if_match_3 = null
        $if_match_4 = null
        $if_match_5 = null
        $if_match_6 = null
        $if_not_match_1 = null
        $if_not_match_2 = null
        $if_not_match_3 = null
        $if_not_match_4 = null
        $if_not_match_5 = null
        $if_not_match_6 = null
        if $ten == 10                    { $if_eq          = true }
        if $unset == null                { $if_eq_null     = true }
        if $ten != 5                     { $if_ne          = true }
        if $ten > 5                      { $if_gt          = true }
        if $ten >= 10                    { $if_gte         = true }
        if $ten < 20                     { $if_lt          = true }
        if $ten <= 10                    { $if_lte         = true }
        if $true                         { $if_true        = true }
        if not $false                    { $if_false       = true }
        if "Steve" in $names             { $if_in          = true }
        if "Bill" not in $names          { $if_not_in      = true }
        if $string =~ /str[aeiou]ng/     { $if_match_1     = true }
        if $string =~ /String/i          { $if_match_2     = true }
        if $string =~ /.*/               { $if_match_3     = true }
        if $string !~ /strong/i          { $if_not_match_1 = true }
        if $string !~ /String/           { $if_not_match_2 = true }
        if $string !~ /^ring$/           { $if_not_match_3 = true }
        if not $string !~ /str[aeiou]ng/ { $if_match_4     = true }
        if not $string !~ /String/i      { $if_match_5     = true }
        if not $string !~ /.*/           { $if_match_6     = true }
        if not $string =~ /strong/i      { $if_not_match_4 = true }
        if not $string =~ /String/       { $if_not_match_5 = true }
        if not $string =~ /^ring$/       { $if_not_match_6 = true }`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestExpressions(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`a`:     2,
		`b`:     6,
		`c`:     20,
		`d`:     5,
		`aa`:    2,
		`bb`:    6,
		`cc`:    20,
		`dd`:    5,
		`f`:     `This 2 is {b} and done`,
		`put_a`: `this is some stuff`,
		`put_b`: "buncha\n    muncha\n    cruncha\n    lines",
	}

	script := `
        $a = 1 + 1
        $b = 9 - 3
        $c = 5 * 4
        $d = 50 / 10
        #$e = 4 * -6 * (3 * 7 + 5) + 2 * 7
        $aa = 1
        $aa += 1
        $bb = 9
        $bb -= 3
        $cc = 5
        $cc *= 4
        $dd = 50
        $dd /= 10
        $f = "This {a}" + ' is {b}' + " and done"

        put """
            this is some stuff
        """ -> $put_a
        put """
            buncha
            muncha
            cruncha
            lines
        """ -> $put_b`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestLoops(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`forevers`:        9,
		`double_break`:    []interface{}{4, 1},
		`double_continue`: []interface{}{8, 9},
		`iterations`:      4,
		`things`: []interface{}{
			float64(1),
			float64(2),
			float64(3),
			float64(4),
			float64(5),
		},
		`topindex`: 9,
		`map`: map[string]interface{}{
			`first`:  float64(1),
			`second`: float64(2),
			`third`:  float64(3),
		},
		`m1`: `first:1`,
		`m2`: `second:2`,
		`m3`: `third:3`,
	}

	script := `
        $forevers = 0
        $double_break = null
        $double_continue = null
        $map = {
            first:  1,
            second: 2,
            third:  3,
        }
        $iterations = null
        $things = [1,2,3,4,5]
        loop {
            if not $index < 10 {
                break
            }
            $forevers = $index
        }
        loop $x in $things {
            $iterations = $index
        }
        loop count 10 {
            $topindex = $index
            loop count 10 {
                if $topindex == 4 {
                    if $index == 2 {
                        break 2
                    }
                }
                $double_break = [$topindex, $index]
            }
        }
        loop count 10 {
            $topindex = $index
            loop count 10 {
                if $topindex == 9 {
                    if $index >= 0 {
                        continue 2
                    }
                }
                $double_continue = [$topindex, $index]
            }
        }

        loop $k, $v in $map {
            if $index == 0 {
                $m1 = "{k}:{v}"
            } else if $index == 1 {
                $m2 = "{k}:{v}"
            } else {
                $m3 = "{k}:{v}"
            }
        }`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestCommands(t *testing.T) {
	assert := require.New(t)

	_, err := eval(`testing::noop`)
	assert.NoError(err)

	actual, err := eval(`testing::noop -> $result`)
	assert.NoError(err)
	assert.Zero(actual[`result`])

	actual, err = eval(`fmt::trim -> $result`)
	assert.NoError(err)
	assert.Zero(actual[`result`])

	actual, err = eval(`fmt::trim "test" {
		prefix: 't',
		suffix: 't',
	} -> $rv`)

	assert.NoError(err)
	assert.Equal(`es`, actual[`rv`])

	actual, err = eval(`fmt::trim "test" -> $rv`)
	assert.NoError(err)
	assert.Equal(`test`, actual[`rv`])
}

func TestHttp(t *testing.T) {
	assert := require.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc(`/json/objects`, func(w http.ResponseWriter, req *http.Request) {
		var into map[string]interface{}

		switch req.Method {
		case `GET`:
			t.Logf("Test HTTP %s: <no body>", req.Method)

			httputil.RespondJSON(w, map[string]interface{}{
				`get`: `got it, good`,
			})
		case `DELETE`:
			httputil.RespondJSON(w, nil)

		case `HEAD`:
			httputil.RespondJSON(w, nil, http.StatusOK)

		default:
			if err := httputil.ParseRequest(req, &into); err == nil {
				t.Logf("Test HTTP %s: %s", req.Method, typeutil.Dump(into))
				httputil.RespondJSON(w, into, http.StatusAccepted)
			} else {
				httputil.RespondJSON(w, err)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	actual, err := eval("http::get %q -> $get_json_object", server.URL+`/json/objects`)
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, maputil.DeepGet(actual[`get_json_object`], []string{`status`}))
	assert.EqualValues(`got it, good`, maputil.DeepGet(actual[`get_json_object`], []string{`body`, `get`}))

	actual, err = eval("http::delete %q -> $delete_json_object", server.URL+`/json/objects`)
	assert.NoError(err)
	assert.EqualValues(http.StatusNoContent, maputil.DeepGet(actual[`delete_json_object`], []string{`status`}))

	actual, err = eval("http::head %q -> $head_json_object", server.URL+`/json/objects`)
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, maputil.DeepGet(actual[`head_json_object`], []string{`status`}))

	actual, err = eval(`http::post %q {
		type: 'json',
		body: {
			test: 1.23,
			data: true,
			value: 'yes',
		},
	} -> $post_json_object`, server.URL+`/json/objects`)

	t.Log(typeutil.Dump(actual[`post_json_object`]))
	assert.NoError(err)
	assert.EqualValues(http.StatusAccepted, maputil.DeepGet(actual[`post_json_object`], []string{`status`}))
	assert.EqualValues(1.23, maputil.DeepGet(actual[`post_json_object`], []string{`body`, `test`}))
	assert.EqualValues(true, maputil.DeepGet(actual[`post_json_object`], []string{`body`, `data`}))
	assert.EqualValues(`yes`, maputil.DeepGet(actual[`post_json_object`], []string{`body`, `value`}))
}

func jsondiff(expected interface{}, actual interface{}) string {
	if expectedJ, err := json.MarshalIndent(expected, ``, `  `); err == nil {
		if actualJ, err := json.MarshalIndent(actual, ``, `  `); err == nil {
			differ := gojsondiff.New()
			if d, err := differ.Compare(expectedJ, actualJ); err == nil {
				formatter := formatter.NewAsciiFormatter(expected, formatter.AsciiFormatterConfig{
					ShowArrayIndex: true,
					Coloring:       true,
				})

				if diffString, err := formatter.Format(d); err == nil {
					return diffString
				} else {
					return fmt.Sprintf("ERROR: formatter: %v", err)
				}
			} else {
				return fmt.Sprintf("ERROR: diff: %v", err)
			}
		} else {
			return fmt.Sprintf("ERROR: actual: %v", err)
		}
	} else {
		return fmt.Sprintf("ERROR: expected: %v", err)
	}
}
