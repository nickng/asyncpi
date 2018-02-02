%{
package asyncpi

import (
	"io"

	"go.nickng.io/asyncpi/internal/name"
)

var proc Process
%}

%union {
	strval string
	proc   Process
	name   Name
	names  []Name
}

%token kLANGLE kRANGLE kLPAREN kRPAREN kPREFIX kSEMICOLON kCOLON kNIL kNAME kREPEAT kNEW kCOMMA
%type <proc> proc simpleproc scope
%type <strval> kNAME
%type <name> scopename
%type <names> names
%type <names> values

%left kPAR
%right kREPEAT
%nonassoc kPREFIX
%right kREP
%left kCOMMA

%%

top : proc { proc = $1 }
    ;

proc :           simpleproc { $$ = $1 }
     | proc kPAR simpleproc { $$ = NewPar($1, $3)}
     ;

simpleproc : kNIL { $$ = NewNilProcess() }
           | kNAME kLANGLE values kRANGLE { $$ = NewSend(name.New($1)); $$.(*Send).SetVals($3) }
           | kNAME kLPAREN names kRPAREN kPREFIX         proc         { $$ = NewRecv(name.New($1), $6); $$.(*Recv).SetVars($3) }
           | kNAME kLPAREN names kRPAREN kPREFIX kLPAREN proc kRPAREN { $$ = NewRecv(name.New($1), $7); $$.(*Recv).SetVars($3) }
           | kLPAREN kNEW scopename kRPAREN scope { $$ = NewRestrict($3, $5) }
           | kLPAREN kNEW scopename kCOMMA names kRPAREN scope { $$ = NewRestricts(append([]Name{$3}, $5...), $7) }
           | kREPEAT         proc { $$ = NewRepeat($2) }
           | kREPEAT kLPAREN proc kRPAREN { $$ = NewRepeat($3) }
           ;

scopename : kNAME              { $$ = name.New($1) }
          | kNAME kCOLON kNAME { $$ = name.NewHinted($1, $3) }
          ;

scope : simpleproc           { $$ = $1 }
      | kLPAREN proc kRPAREN { $$ = $2 }
      ;

names : /* empty */            { $$ = nil }
      |              scopename { $$ = []Name{$1} }
      | names kCOMMA scopename { $$ = append($1, $3) }
      ;

values : /* empty */         { $$ = nil }
       |               kNAME { $$ = []Name{name.New($1)} }
       | values kCOMMA kNAME { $$ = append($1, name.New($3)) }
       ;

%%

// Parse is the entry point to the asyncpi calculus parser.
func Parse(r io.Reader) (Process, error) {
	l := newLexer(r)
	asyncpiParse(l)
	select {
	case err := <-l.Errors:
		return nil, err
	default:
		return proc, nil
	}
}
