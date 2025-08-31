package chisel

import (
	"flag"
)

type Args struct {
	ChiselFilePath              string
	Language                    string
	Directory                   string
	ProjectDirectory            string
	Overwrite                   bool
	GeneratedLibDirectory       string
	CopyLibrary                 bool
	GeneratedConstructDirectory string
}

func Init() Args {
	args := Args{}
	flag.StringVar(&args.Language, "l", "cpp", "The language (default=cpp).")
	flag.StringVar(&args.Directory, "d", ".", "The directory to write to (default=.).")
	flag.StringVar(&args.ProjectDirectory, "p", ".", "The project directory (default=.).")
	flag.StringVar(&args.GeneratedLibDirectory, "g", ".", "Where to output protected generated library files relative to -d directory. (default=.).")
	flag.StringVar(&args.GeneratedConstructDirectory, "G", ".", "Where to output user editable construct classes relative to -d directory. (default=.).")
	flag.BoolVar(&args.CopyLibrary, "copy-lib", false, "Whether to copy the library files or not. (default=false).")
	flag.BoolVar(&args.Overwrite, "overwrite", false, "Overwrite existing files without asking user (default=false).")
	flag.Parse()

	args.ChiselFilePath = flag.Arg(0)
	return args
}
