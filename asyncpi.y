%{
package asyncpi

import (
	"io"
)

var proc Process
%}

%union {
	strval string
	proc   Process
	name   Name
	names  []Name
}

%token LANGLE RANGLE LPAREN RPAREN PREFIX SEMICOLON COLON NIL NAME REPEAT NEW COMMA
%type <proc> proc simpleproc scope
%type <strval> NAME
%type <name> scopename
%type <names> names
%type <names> values

%left PAR
%right REPEAT
%nonassoc PREFIX
%right REP
%left COMMA

%%

top : proc { proc = $1 }
    ;

proc :          simpleproc { $$ = $1 }
     | proc PAR simpleproc { $$ = NewPar($1, $3)}
     ;

simpleproc : NIL { $$ = NewNilProcess() }
           | NAME LANGLE values RANGLE { $$ = NewSend(newPiName($1)); $$.(*Send).SetVals($3) }
           | NAME LPAREN names RPAREN PREFIX proc { $$ = NewRecv(newPiName($1), $6); $$.(*Recv).SetVars($3) }
           | NAME LPAREN names RPAREN PREFIX LPAREN proc RPAREN { $$ = NewRecv(newPiName($1), $7); $$.(*Recv).SetVars($3) }
           | LPAREN NEW scopename RPAREN scope { $$ = NewRestrict($3, $5) }
           | LPAREN NEW scopename COMMA names RPAREN scope { $$ = NewRestricts(append([]Name{$3}, $5...), $7) }
           | REPEAT proc { $$ = NewRepeat($2) }
           ;

scopename : NAME            { $$ = newPiName($1) }
          | NAME COLON NAME { $$ = newTypedPiName($1, $3) }

scope : simpleproc         { $$ = $1 }
      | LPAREN proc RPAREN { $$ = $2 }
      ;

names : /* empty */           { $$ = []Name{} }
      |             scopename { $$ = []Name{$1} }
      | names COMMA scopename { $$ = append($1, $3) }
      ;

values : /* empty */       { $$ = []Name{} }
       | NAME              { $$ = []Name{newPiName($1)} }
       | values COMMA NAME { $$ = append($1, newPiName($3)) }
       ;

%%

// Parse is the entry point to the asyncpi calculus parser.
func Parse(r io.Reader) (Process, error) {
	l := NewLexer(r)
	asyncpiParse(l)
	select {
	case err := <-l.Errors:
		return nil, err
	default:
		return proc, nil
	}
}
