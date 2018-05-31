package builtin

import (
	"bytes"
	"testing"

	"github.com/brettbuddin/shaden/errors"
	"github.com/brettbuddin/shaden/lisp"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	errors.Separator = ": "

	var tests = []struct {
		input  []byte
		result interface{}
		error  string
	}{
		// Data types
		{input: []byte("1"), result: 1},
		{input: []byte(`"hello"`), result: `hello`},
		{input: []byte(`(string-split "hello/world" "/")`), result: lisp.List{"hello", "world"}},
		{input: []byte(`(string-join (list "hello" "world") "/")`), result: "hello/world"},
		{input: []byte(`(string-has-prefix "hello-world" "hello-")`), result: true},
		{input: []byte(`(string-replace "hi-world" "world" "mark" -1)`), result: "hi-mark"},
		{input: []byte(`(string-replace "hi-world-world" "world" "mark" 1)`), result: "hi-mark-world"},
		{input: []byte(`(string? (string :hello))`), result: true},
		{input: []byte(`:hello`), result: lisp.Keyword(`hello`)},
		{input: []byte(`(keyword "hello")`), result: lisp.Keyword(`hello`)},
		{input: []byte(`(keyword (keyword "hello"))`), result: lisp.Keyword(`hello`)},
		{input: []byte(`false`), result: false},
		{input: []byte(`true`), result: true},
		{input: []byte(`nil`), result: nil},
		{input: []byte(`(nil? nil)`), result: true},
		{input: []byte(`(nil? 1)`), result: false},
		{input: []byte(`(number? 1)`), result: true},
		{input: []byte(`(number? "hello")`), result: false},
		{input: []byte(`(boolean? "hello")`), result: false},
		{input: []byte(`(boolean? false)`), result: true},
		{input: []byte(`(fn? false)`), result: false},
		{input: []byte(`(fn? (table))`), result: true},
		{input: []byte(`(fn? (list))`), result: true},
		{input: []byte(`(fn? :name)`), result: true},
		{input: []byte(`(fn? "name")`), result: false},
		{input: []byte(`(fn? (fn () (+ 1 1)))`), result: true},
		{input: []byte(`(define x 1) (symbol? x)`), result: false},
		{input: []byte(`(define x 1) (defined? "x")`), result: true},
		{input: []byte(`(define x 1) (defined? "y")`), result: false},
		{input: []byte(`(symbol? (symbol "x"))`), result: true},
		{input: []byte(`(symbol? (quote x))`), result: true},
		{input: []byte(`(keyword? (keyword "hello"))`), result: true},
		{input: []byte(`(list? (list))`), result: true},
		{input: []byte(`(list? 1)`), result: false},
		{input: []byte(`(error? (errorf "hello"))`), result: true},
		{input: []byte(`(int? 1)`), result: true},
		{input: []byte(`(int? "nope")`), result: false},
		{input: []byte(`(int? 1.0)`), result: false},
		{input: []byte(`(float? 1.0)`), result: true},
		{input: []byte(`(float? "nope")`), result: false},
		{input: []byte(`(float? 1)`), result: false},
		{input: []byte(`(table? "abcd")`), result: false},
		{input: []byte(`(table? (table))`), result: true},
		{input: []byte(`(table? (quote (table)))`), result: false},
		{input: []byte(`(empty? "abcd")`), result: false},
		{input: []byte(`(empty? "")`), result: true},
		{input: []byte(`(empty? (table :a :b))`), result: false},
		{input: []byte(`(empty? (table))`), result: true},
		{input: []byte(`(empty? (list 1 2 3))`), result: false},
		{input: []byte(`(empty? (list))`), result: true},
		{input: []byte(`(list? (quote (table)))`), result: true},
		{input: []byte(`(eval (quote (+ 1 1)))`), result: 2},
		{input: []byte(`(eval (read "(+ 1 1)"))`), result: 2},
		{input: []byte(`(eval (quasiquote (list 1 2 3)))`), result: lisp.List{1, 2, 3}},
		{input: []byte(`(define x 5) (eval (quasiquote (list 1 (unquote (+ 2 x)) 3)))`), result: lisp.List{1, 7, 3}},
		{input: []byte(`(define-macro (infix infixed) (list (infixed 1) (infixed 0) (infixed 2))) (infix (1 + 3))`), result: 4},

		// Collections
		{input: []byte(`(list 1 2 3)`), result: lisp.List{1, 2, 3}},
		{input: []byte(`(len (list 1 2 3))`), result: 3},
		{input: []byte(`((list 1 2 3) 1)`), result: 2},
		{input: []byte(`(table-get (table :a 1) :a)`), result: 1},
		{input: []byte(`(:a (table :a 1))`), result: 1},
		{input: []byte(`((table :a 1) :a)`), result: 1},
		{input: []byte(`(table-exists? (table :a 1) :a)`), result: true},
		{input: []byte(`(table-exists? (table :a 1) :b)`), result: false},
		{input: []byte(`(let ((hm (table :a 1))) (table-set hm :a 3) (table-get hm :a))`), result: 3},
		{input: []byte(`(let ((hm (table :a 1))) (table-delete hm :a) (table-get hm :a))`), result: nil},
		{input: []byte(`(rest (list 1 2 3))`), result: lisp.List{2, 3}},
		{input: []byte(`(first (list 1 2 3))`), result: 1},
		{input: []byte(`(cons 1 (list 2 3 4))`), result: lisp.List{1, 2, 3, 4}},
		{input: []byte(`(append (list 1 2 3) 4 5 6)`), result: lisp.List{1, 2, 3, 4, 5, 6}},
		{input: []byte(`(prepend (list 1 2 3) 4 5 6)`), result: lisp.List{4, 5, 6, 1, 2, 3}},
		{input: []byte(`(table :hello "world")`), result: lisp.Table{lisp.Keyword("hello"): "world"}},
		{input: []byte(`(table-merge (table :hello "world") (table :world "hello"))`), result: lisp.Table{
			lisp.Keyword("hello"): "world",
			lisp.Keyword("world"): "hello",
		}},
		{input: []byte(`(table-select (table :hello "world" :world "hello") (fn (k _) (= k :hello)))`), result: lisp.Table{
			lisp.Keyword("hello"): "world",
		}},
		{input: []byte(`(reduce (fn (r i v) (+ r v)) 0 (list 1 2 3))`), result: 6},
		{input: []byte(`(reduce (fn (r k v) (+ r v)) 0 (table :a 2 :b 3))`), result: 5},

		// Math
		{input: []byte(`(+ 1 1)`), result: 2},
		{input: []byte(`(+ 1.0 1)`), result: 2.0},
		{input: []byte(`(- 3 1)`), result: 2},
		{input: []byte(`(- 3.0 1)`), result: 2.0},
		{input: []byte(`(* 2 2)`), result: 4},
		{input: []byte(`(* 2.0 2)`), result: 4.0},
		{input: []byte(`(/ 8 2)`), result: 4},
		{input: []byte(`(/ 8.0 2)`), result: 4.0},
		{input: []byte(`(pow 2 3)`), result: float64(8)},
		{input: []byte(`(pow 2.0 3)`), result: float64(8)},
		{input: []byte(`(float 2)`), result: float64(2)},
		{input: []byte(`(float 2.0)`), result: float64(2)},
		{input: []byte(`(int 2.34)`), result: int(2)},
		{input: []byte(`(int 2)`), result: int(2)},

		// Conditionals
		{input: []byte(`(= 1 1)`), result: true},
		{input: []byte(`(= 1 "a")`), result: false},
		{input: []byte(`(!= 1 1)`), result: false},
		{input: []byte(`(!= 1 "a")`), result: true},
		{input: []byte(`(> 1 2)`), result: false},
		{input: []byte(`(> 2 1)`), result: true},
		{input: []byte(`(> 2 2)`), result: false},
		{input: []byte(`(> 2.0 2)`), error: "failed to call >: cannot compare float64 and int"},
		{input: []byte(`(< 1 2)`), result: true},
		{input: []byte(`(< 2 1)`), result: false},
		{input: []byte(`(< 2 2)`), result: false},
		{input: []byte(`(< 2.0 2)`), error: "failed to call <: cannot compare float64 and int"},
		{input: []byte(`(not (= 1 2))`), result: true},
		{input: []byte(`(if true "hello" "world")`), result: "hello"},
		{input: []byte(`(if false "hello" "world")`), result: "world"},
		{input: []byte(`(if nil "hello" "world")`), result: "world"},
		{input: []byte(`(if 5 "hello" "world")`), result: "hello"},
		{input: []byte(`(let ((x 1)) 
							 (cond ((string? x) "string") 
								   ((number? x) "number")))`), result: "number"},
		{input: []byte(`(when true "hello")`), result: "hello"},
		{input: []byte(`(when false "hello")`), result: nil},
		{input: []byte(`(when nil "hello")`), result: nil},
		{input: []byte(`(when 5 "hello")`), result: "hello"},
		{input: []byte(`(unless true "hello")`), result: nil},
		{input: []byte(`(unless false "hello")`), result: "hello"},
		{input: []byte(`(unless nil "hello")`), result: "hello"},
		{input: []byte(`(unless 5 "hello")`), result: nil},
		{input: []byte(`(or)`), result: false},
		{input: []byte(`(or 1)`), result: 1},
		{input: []byte(`(or 1 1)`), result: true},
		{input: []byte(`(or nil 1)`), result: true},
		{input: []byte(`(or false 1)`), result: true},
		{input: []byte(`(and)`), result: true},
		{input: []byte(`(and 1)`), result: 1},
		{input: []byte(`(and 1 2)`), result: true},
		{input: []byte(`(and 1 nil)`), result: false},
		{input: []byte(`(and 1 false)`), result: false},
		{input: []byte(`(and false 1)`), result: false},

		// Definitions and Functions
		{input: []byte(`(define hello 100) hello`), result: 100},
		{input: []byte(`((fn (x y) (+ x y)) 5 8)`), result: 13},
		{input: []byte(`((fn (x y) (set x (+ x y)) (+ x 1)) 5 8)`), result: 14},
		{input: []byte(`((fn (_ y) _) 1 2)`), error: "failed to call anonymous function: undefined symbol _"},
		{input: []byte(`((fn (_ y) y) 1 2)`), result: 2},
		{input: []byte(`(apply (fn (x y) (+ x y)) (list 5 8))`), result: 13},
		{input: []byte(`(apply (fn (x y) (+ x y)) 5 9)`), result: 14},
		{input: []byte(`(define (add1 x) (+ x 1)) (add1 1)`), result: 2},
		{input: []byte(`(define (add3 x) (set x (+ x 1)) (+ x 2)) (add3 1)`), result: 4},
		{input: []byte(`(do (/ 10 2) (* 2 2))`), result: 4},
		{input: []byte(`(let ((x 3) (y 4)) (* x y))`), result: 12},

		// Variadic Functions
		{input: []byte(`((fn (x & y) y) 1 2 3 4 5)`), result: lisp.List{2, 3, 4, 5}},
		{input: []byte(`(define (vary x y & z) x) (vary 1 2 3 4 5)`), result: 1},
		{input: []byte(`(define (vary x y & z) y) (vary 1 2 3 4 5)`), result: 2},
		{input: []byte(`(define (vary x y & z) z) (vary 1 2 3 4 5)`), result: lisp.List{3, 4, 5}},
		{input: []byte(`(define (vary x y & z w) z) (vary 1 2 3 4 5)`), error: "failed to call vary: definition has too many arguments after variadic symbol &"},
		{input: []byte(`(define (vary x y & z) z) (vary 1 2)`), result: lisp.List{}},

		// Iterators
		{input: []byte(`(map (fn (i v) (+ 1 v)) (list 1 2 3))`), result: lisp.List{2, 3, 4}},
		// {input: []byte(`(map (fn (i v) (+ 1 v)) (table :one 1 :two 2 :three 3))`), result: lisp.List{2, 3, 4}},
		{input: []byte(`(each (fn (k v) (+ 1 v)) (table :a 1 :b 2))`), result: lisp.Table{
			lisp.Keyword("a"): 1,
			lisp.Keyword("b"): 2,
		}},
		{input: []byte(`(each (fn (i v) (+ 1 v)) (list 1 2 3))`), result: lisp.List{1, 2, 3}},
		{input: []byte(`(define x 0) (dotimes (i 3) (set x (+ x 1))) x`), result: 3},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			node, err := lisp.Parse(bytes.NewBuffer(test.input))
			require.NoError(t, err)

			env := lisp.NewEnvironment()
			Load(env)
			result, err := env.Eval(node)

			if test.error != "" {
				require.Error(t, err)
				require.Equal(t, test.error, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.result, result)
		})
	}
}

func TestLoad(t *testing.T) {
	node, err := lisp.Parse(bytes.NewBufferString(`(load "testdata/load-example.lisp")`))
	require.NoError(t, err)

	env := lisp.NewEnvironment()
	Load(env)
	result, err := env.Eval(node)
	require.NoError(t, err)
	require.Equal(t, 2, result)
}

func TestLoad_CustomLoadPath(t *testing.T) {
	node, err := lisp.Parse(bytes.NewBufferString(`(set load-path (list "testdata/load-path/")) (load "load-path-example.lisp")`))
	require.NoError(t, err)

	env := lisp.NewEnvironment()
	Load(env)
	result, err := env.Eval(node)
	require.NoError(t, err)
	require.Equal(t, 2, result)
}

func TestLoad_UnknownFile(t *testing.T) {
	node, err := lisp.Parse(bytes.NewBufferString(`(load "testdata/doesntexist.lisp")`))
	require.NoError(t, err)

	env := lisp.NewEnvironment()
	Load(env)
	_, err = env.Eval(node)
	require.Error(t, err)
}
