package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const MB = 1 << 20

//go:embed index.html
var indexFile []byte

type function struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

var registeredFunctions = map[string]*function{}

func makeLogI32(modname string) func(uint32) {
	return func(v uint32) {
		log.Printf("[%v]: %v", modname, v)
	}
}

func makeLogString(modname string) func(context.Context, api.Module, uint32, uint32) {
	return func(_ context.Context, mod api.Module, ptr, len uint32) {
		buf, ok := mod.Memory().Read(ptr, len)
		if ok {
			log.Printf("[%v]: %v", modname, string(buf))
		} else {
			log.Printf("[%v]: failed to read string from memory (%v, %v)",
				modname, ptr, len)
		}
	}
}

func registerFunction(ctx context.Context, modname string, code []byte) error {
	if registeredFunctions == nil {
		registeredFunctions = make(map[string]*function, 1)
	}

	r := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(makeLogI32(modname)).Export("log_i32").
		NewFunctionBuilder().WithFunc(makeLogString(modname)).Export("log_string").
		Instantiate(ctx)
	if err != nil {
		return err
	}

	log.Printf("compiling module %v", modname)

	cm, err := r.CompileModule(ctx, code)
	if err != nil {
		log.Printf("failed: %v", err)

		return err
	}

	fn := &function{
		runtime: r,
		module:  cm,
	}

	registeredFunctions[modname] = fn

	return nil
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	modname := vars["modname"]

	log.Printf("function %v requested with query %v", modname, r.URL.Query())

	fn, ok := registeredFunctions[modname]
	if !ok {
		err := fmt.Errorf("function %q is not registered", modname)

		log.Print(err)
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	env := map[string]string{
		"http_path":   r.URL.Path,
		"http_method": r.Method,
		"http_host":   r.Host,
		"http_query":  r.URL.Query().Encode(),
		"remote_addr": r.RemoteAddr,
	}

	// Set up stdout redirection and env vars for the module.
	var stdoutBuf bytes.Buffer

	config := wazero.NewModuleConfig().WithStdout(&stdoutBuf)

	for k, v := range env {
		config = config.WithEnv(k, v)
	}

	// Instantiate the module. This invokes the _start function by default.
	if _, err := fn.runtime.InstantiateModule(r.Context(), fn.module, config); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	fmt.Fprint(w, stdoutBuf.String())
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write(indexFile)
	case http.MethodPost:
		if err := r.ParseMultipartForm(10 * MB); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		modname := r.FormValue("modname")
		if modname == "" {
			log.Print("no module name provided")
			http.Error(w, "no module name provided", http.StatusBadRequest)

			return
		}

		file, _, err := r.FormFile("modfile")
		if err != nil {
			log.Printf("failed to read %v: %v", modname, err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		defer file.Close()

		code, err := io.ReadAll(file)
		if err != nil {
			log.Printf("failed to read %v: %v", modname, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if err := registerFunction(r.Context(), modname, code); err != nil {
			log.Printf("failed to register %v: %v", modname, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		fmt.Fprintf(w, "registered /run/%v", modname)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", registerHandler)
	r.HandleFunc("/run/{modname}", runHandler)

	log.Println("Started on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}
