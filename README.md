# ssaview 

tool for hacking the detail of the SSA in Golang compiler.


# Usage:
$ go get github.com/Katsusan/ssaview    
$ ssaview -f=main -args="-N -l" -h=127.0.0.1 -p=9000 hello.go   
$ ssaview -f=main hello.go   
$ ssaview hello.go world.go   

# notes:
There are 49 steps from AST to final assembly code.
all define in [ssa source](https://github.com/golang/go/blob/5337e53dfa3f5fde73b8f505ec3a91c628e8f648/src/cmd/compile/internal/ssa/compile.go#L459-L513).

That is, 53 columns are shown in default.
If you want to generate SVG of those 49 steps, please use `-f myfunc:x-y,z`, in which x, y, z are ssa passes name.
for example, `-f main:number_lines-early_copyelim,stackfram`. ("_" will be replaced by space when parsed)

```
"sources"
"AST"
"start"
"number lines"      //ssa passes[0]
"early phielim"     //ssa passes[1]
"early copyelim"
...
"stackframe"        //ssa passes[47]
"trim"              //ssa passes[48]
"genssa"            //final assembly
```

# Preview
![](/ssa_web.png)
