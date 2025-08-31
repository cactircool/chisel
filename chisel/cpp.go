package chisel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type UnitData struct {
	ClassName    string
	ClassCreated bool
	Ref          Unit
}

var unitInfo = map[string]UnitData{}
var customUnitFileCreated = false

func GenerateCppControlFlow(constructs []*Construct, args Args) error {
	file, err := os.Create(filepath.Join(args.Directory, args.GeneratedLibDirectory, "Reader.hpp"))
	if err != nil {
		return err
	}
	defer file.Close()

	includes := []string{}
	data := func() string {
		var sb strings.Builder
		for _, construct := range constructs {
			includes = append(includes, (fmt.Sprintf("#include \"Buffered%sReader.hpp\"", construct.Name)))
			sb.WriteString(fmt.Sprintf("chisel::Buffered%sReader %s() { return chisel::Buffered%sReader(this->_reader); }\n", construct.Name, construct.Name, construct.Name))
		}
		return sb.String()
	}()

	g := guard()
	_, err = fmt.Fprintf(file, `
#ifndef %s
#define %s

#include "chisel/chisel.hpp"
%s

namespace chisel {

class Reader {
	chisel::ChiselIstream &_reader;
public:
	Reader(chisel::ChiselIstream &reader) : _reader(reader) {}

	%s
};

}

#endif // %s
`, g, g, strings.Join(includes, "\n"), data, g)

	if err != nil {
		return err
	}
	return nil
}

func GenerateCppFilesFromConstruct(construct *Construct, args Args) error {
	err := generateConstruct(construct, args)
	if err != nil {
		return err
	}

	err = generateBufferedReader(construct, args)
	if err != nil {
		return err
	}

	err = generateUnits(construct, args)
	if err != nil {
		return err
	}

	if construct.Root {
		err = generateRootFlow(construct, args)
		if err != nil {
			return err
		}
	}
	return nil
}

func guard() string {
	return fmt.Sprint("A_", strings.ReplaceAll(uuid.NewString(), "-", "_"))
}

func generateRootFlow(construct *Construct, args Args) error {
	flowPath := filepath.Join(args.Directory, args.GeneratedConstructDirectory, "flow.hpp")
	fmt.Println("Flow path:", flowPath)
	file, err := os.Create(flowPath)
	if err != nil {
		return err
	}
	defer file.Close()

	g := guard()
	_, err = fmt.Fprintf(file, `
#ifndef %s
#define %s

#include "chisel/chisel.hpp"
#include "%s.hpp"
#include "Buffered%sReader.hpp"

namespace chisel {

	std::unique_ptr<%s> parse(chisel::ChiselIstream &stream) {
		Buffered%sReader reader(stream);
		return std::make_unique<%s>(reader);
	}

}

#endif // %s
`, g, g, construct.Name, construct.Name, construct.Name, construct.Name, construct.Name, g)

	if err != nil {
		return err
	}
	return nil
}

func generateUnits(construct *Construct, args Args) error {
	file, err := os.OpenFile(filepath.Join(args.Directory, args.GeneratedLibDirectory, "units.hpp"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if !customUnitFileCreated {
		fmt.Fprint(file, `
#include "chisel/chisel.hpp"

#include <memory>
#include <iostream>
#include <cstring>
#include <cstdlib>
`)
		customUnitFileCreated = true
	}

	for _, unit := range construct.Units {
		q := []*UnitReferenceTree{unit.ReferenceTree()}
		for len(q) > 0 {
			front := q[0]
			q = append(q, front.Children...)
			q = q[1:]
			_ = fetchUnitData(front.Ref)
		}
	}

	g := guard()
	if _, err = fmt.Fprintln(file, "#ifndef ", g); err != nil {
		return err
	}
	if _, err = fmt.Fprintln(file, "#define ", g); err != nil {
		return err
	}
	if _, err = fmt.Fprintln(file, "namespace chisel {"); err != nil {
		return err
	}

	for _, data := range unitInfo {
		fmt.Fprintf(file, "class %s;\n", data.ClassName)
	}
	if _, err = fmt.Fprintln(file, "}"); err != nil {
		return err
	}

	sources := make([]string, 0, len(unitInfo))
	for _, data := range unitInfo {
		unit := data.Ref
		unitData := fetchUnitData(unit)
		if unitData.ClassCreated {
			continue
		}

		unitData.ClassCreated = true
		def, source, err := generateUnitContent(unit)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(file, "namespace chisel {\n%s\n}\n", def)
		if err != nil {
			return err
		}

		sources = append(sources, source)
	}

	for _, source := range sources {
		fmt.Fprintf(file, "namespace chisel {\n%s\n}\n", source)
	}

	if _, err = fmt.Fprintln(file, "#endif // ", g); err != nil {
		return err
	}
	return nil
}

func generateUnitContent(unit Unit) (string, string, error) {
	switch v := unit.(type) {
	case BasicRegex:
		return generatBasicRegexContent(&v)
	case *BasicRegex:
		return generatBasicRegexContent(v)

	case RangeRegex:
		return generateRangeRegexContent(&v)
	case *RangeRegex:
		return generateRangeRegexContent(v)

	case OptionRegex:
		return generateOptionRegexContent(&v)
	case *OptionRegex:
		return generateOptionRegexContent(v)

	case OptionalRegex:
		return generateOptionalRegexContent(&v)
	case *OptionalRegex:
		return generateOptionalRegexContent(v)

	case StarRegex:
		return generateStarRegexContent(&v)
	case *StarRegex:
		return generateStarRegexContent(v)

	case PlusRegex:
		return generatePlusRegexContent(&v)
	case *PlusRegex:
		return generatePlusRegexContent(v)

	case OrRegex:
		return generateOrRegexContent(&v)
	case *OrRegex:
		return generateOrRegexContent(v)

	case LiteralUnit:
		return generateLiteralUnitContent(&v)
	case *LiteralUnit:
		return generateLiteralUnitContent(v)

	case ReferenceUnit:
		return generateReferenceUnitContent(&v)
	case *ReferenceUnit:
		return generateReferenceUnitContent(v)

	default:
		return "", "", fmt.Errorf("Unknown/unimplemented unit type!")
	}
}

func generatBasicRegexContent(b *BasicRegex) (string, string, error) {
	unitData := fetchUnitData(b)

	data := map[string]any{
		"Name":  unitData.ClassName,
		"Match": byte(b.Match),
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _single_read;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	chisel::BufferedData read() override;
	void reset() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _single_read(false) {}
void {{.Name}}::reset() {
	_single_read = false;
}
chisel::BufferedData {{.Name}}::read() {
	char c = this->reader().get();
	if (c == {{.Match}} && !this->_single_read) {
		this->_single_read = true;
		return { .data = c };
	}

	this->reader().putback(c);
	if (this->_single_read)
		return { .finished = true };
	return { .failed = true };
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateRangeRegexContent(r *RangeRegex) (string, string, error) {
	unitData := fetchUnitData(r)

	data := map[string]any{
		"Name": unitData.ClassName,
		"From": r.From,
		"To":   r.To,
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _single_read;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _single_read(false) {}
void {{.Name}}::reset() {
	_single_read = false;
}
chisel::BufferedData {{.Name}}::read() {
	char c = this->reader().get();
	if (c >= {{.From}} && c <= {{.To}} && !this->_single_read) {
		this->_single_read = true;
		return { .data = c };
	}

	this->reader().putback(c);
	if (this->_single_read)
		return { .finished = true };
	return { .failed = true };
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateOptionRegexContent(o *OptionRegex) (string, string, error) {
	unitData := fetchUnitData(o)
	matchData := []UnitData{}
	for _, option := range o.Options {
		matchData = append(matchData, fetchUnitData(option))
	}

	data := map[string]any{
		"Name": unitData.ClassName,
		"ReadBody": func() string {
			var sb strings.Builder
			sb.WriteString("chisel::BufferedData data;\n")
			for i, data := range matchData {
				sb.WriteString(fmt.Sprintf("static %s u%d(reader());\n", data.ClassName, i))
				sb.WriteString(fmt.Sprintf("u%d.reset();\n", i))
				sb.WriteString(fmt.Sprintf("data = u%d.read();\n", i))
				sb.WriteString("if (!data.failed && !data.finished) { _single_read = true; return data; }\n")
			}
			sb.WriteString("_single_read = true; return { .failed = true };\n")
			return sb.String()
		}(),
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _single_read;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _single_read(false) {}
void {{.Name}}::reset() {
	_single_read = false;
}
chisel::BufferedData {{.Name}}::read() {
	if (_single_read)
		return { .finished = true };
	{{.ReadBody}}
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateOptionalRegexContent(o *OptionalRegex) (string, string, error) {
	unitData := fetchUnitData(o)
	matchData := fetchUnitData(o.Match)

	data := map[string]any{
		"Name":      unitData.ClassName,
		"UnitClass": matchData.ClassName,
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _single_read;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _single_read(false) {}
void {{.Name}}::reset() {
	_single_read = false;
}
chisel::BufferedData {{.Name}}::read() {
	static {{.UnitClass}} unit(reader());
	auto data = unit.read();
	_single_read = true;
	if (data.failed)
		return { .finished = true };
	return data;
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateStarRegexContent(s *StarRegex) (string, string, error) {
	unitData := fetchUnitData(s)
	matchData := fetchUnitData(s.Match)

	data := map[string]any{
		"Name":      unitData.ClassName,
		"UnitClass": matchData.ClassName,
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _completed;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _completed(false) {}
void {{.Name}}::reset() {
	_completed = false;
}
chisel::BufferedData {{.Name}}::read() {
	static {{.UnitClass}} unit(reader());
	if (_completed)
		return { .finished = true };
	while (true) {
		auto data = unit.read();
		if (data.finished) {
			unit.reset();
			continue;
		}
		if (data.failed) {
			this->_completed = true;
			return { .finished = true };
		}
		return data;
	}
	return { .failed = true };
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	_s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = _s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generatePlusRegexContent(p *PlusRegex) (string, string, error) {
	unitData := fetchUnitData(p)
	matchData := fetchUnitData(p.Match)

	data := map[string]any{
		"Name":      unitData.ClassName,
		"UnitClass": matchData.ClassName,
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	bool _completed, _single_read;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader), _completed(false), _single_read(false) {}
void {{.Name}}::reset() {
	_completed = false;
	_single_read = false;
}
chisel::BufferedData {{.Name}}::read() {
	static {{.UnitClass}} unit(reader());
	while (true) {
		auto data = unit.read();
		if (data.finished) {
			unit.reset();
			continue;
		}
		if (data.failed) {
			this->_completed = true;
			return this->_single_read ? chisel::BufferedData{ .finished = true } : chisel::BufferedData{ .failed = true };
		}
		this->_single_read = true;
		return data;
	}
	return { .failed = true };
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	_s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = _s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateOrRegexContent(o *OrRegex) (string, string, error) {
	unitData := fetchUnitData(o)

	data := map[string]any{
		"Name":          unitData.ClassName,
		"OptionsLength": len(o.Options),
		"OptionsArray": func() string {
			var s strings.Builder
			for i, option := range o.Options {
				optionData := fetchUnitData(option)
				s.WriteString(fmt.Sprintf("units[%d] = new %s(this->reader());\n", i, optionData.ClassName))
			}
			return s.String()
		}(),
	}

	defs := `
class {{.Name}} : public chisel::OrUnit {
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	std::unique_ptr<Unit> extract_option() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader) {}
void {{.Name}}::reset() {}
std::unique_ptr<Unit> {{.Name}}::extract_option() {
	Unit *units[{{.OptionsLength}}];
	{{.OptionsArray}}

	size_t valid_units = {{.OptionsLength}};
	std::stringstream buffer;
	Unit *ret = nullptr;

	while (true) {
		for (size_t i = 0; i < {{.OptionsLength}}; ++i) {
			if (!units[i]) continue;

			auto data = units[i]->read();
			if (data.failed) {
				--valid_units;
				units[i] = nullptr;
			} else if (!data.finished) {
				if (i < {{.OptionsLength}} - 1) {
					this->reader().putback(data.data);
				} else {
					buffer << data.data;
				}
			}
		}

		if (valid_units == 1) {
			for (size_t i = 0; i < {{.OptionsLength}}; ++i)
				if (units[i]) {
					ret = units[i];
					break;
				}
		}

		if (valid_units == 0)
			break;
	}

	for (size_t i = 0; i < {{.OptionsLength}}; ++i)
		delete units[i];

	this->reader().prepend_stream(buffer);
	return std::unique_ptr<Unit>(ret.unit);
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	_s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = _s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateLiteralUnitContent(l *LiteralUnit) (string, string, error) {
	unitData := fetchUnitData(l)

	data := map[string]any{
		"Name":    unitData.ClassName,
		"Literal": l.Literal,
	}

	defs := `
class {{.Name}} : public chisel::Unit {
	static const char *_literal;
	static size_t _index;
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	void reset() override;
	chisel::BufferedData read() override;
};
`

	sources := `
const char *{{.Name}}::_literal = "{{.Literal}}";
size_t {{.Name}}::_index = 0;
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader) {}
void {{.Name}}::reset() {
	_index = 0;
}
chisel::BufferedData {{.Name}}::read() {
	char c = this->reader().get();
	if (_index >= strlen(_literal)) {
		this->reader().putback(c);
		return { .finished = true };
	}
	if (c == _literal[_index++])
		return { .data = c };
	this->reader().putback(c);
	return { .failed = true };
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	_s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = _s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func generateReferenceUnitContent(r *ReferenceUnit) (string, string, error) {
	unitData := fetchUnitData(r)

	data := map[string]any{
		"Name":          unitData.ClassName,
		"ConstructName": r.Reference,
	}

	defs := `
class {{.Name}} : public chisel::ReferenceUnit {
public:
	{{.Name}}(chisel::ChiselIstream &reader);
	std::unique_ptr<BufferedConstructReader> extract_construct() override;
};
`

	sources := `
{{.Name}}::{{.Name}}(chisel::ChiselIstream &reader) : chisel::Unit(reader) {}
std::unique_ptr<BufferedConstructReader> {{.Name}}::extract_construct() {
	return std::make_unique<Buffered{{.ConstructName}}Reader>(this->reader());
}
`

	var sb strings.Builder
	var sc strings.Builder
	d := template.Must(template.New("").Parse(defs))
	_s := template.Must(template.New("").Parse(sources))
	err := d.Execute(&sb, data)
	if err != nil {
		return "", "", err
	}
	err = _s.Execute(&sc, data)
	return sb.String(), sc.String(), nil
}

func fetchUnitData(unit Unit) UnitData {
	val, ok := unitInfo[unit.Name()]
	if ok {
		return val
	}
	data := UnitData{
		ClassName: fmt.Sprint(func() string {
			switch unit.(type) {
			case BasicRegex:
				return "BasicRegex"
			case *BasicRegex:
				return "BasicRegex"

			case RangeRegex:
				return "RangeRegex"
			case *RangeRegex:
				return "RangeRegex"

			case OptionRegex:
				return "OptionRegex"
			case *OptionRegex:
				return "OptionRegex"

			case OptionalRegex:
				return "OptionalRegex"
			case *OptionalRegex:
				return "OptionalRegex"

			case StarRegex:
				return "StarRegex"
			case *StarRegex:
				return "StarRegex"

			case PlusRegex:
				return "PlusRegex"
			case *PlusRegex:
				return "PlusRegex"

			case OrRegex:
				return "OrRegex"
			case *OrRegex:
				return "OrRegex"

			case LiteralUnit:
				return "Literal"
			case *LiteralUnit:
				return "Literal"

			case ReferenceUnit:
				return "Reference"
			case *ReferenceUnit:
				return "Reference"

			default:
				log.Fatal("Unknown/unimplemented unit type!")
				return ""
			}
		}(), "Unit_"+strings.ReplaceAll(uuid.NewString(), "-", "_")),
		Ref: unit,
	}
	unitInfo[unit.Name()] = data
	return data
}

func generateBufferedReader(construct *Construct, args Args) error {
	file, err := os.Create(filepath.Join(args.Directory, args.GeneratedLibDirectory, fmt.Sprintf("Buffered%sReader.hpp", construct.Name)))
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]any{
		"Guard":       guard(),
		"Name":        fmt.Sprintf("Buffered%sReader", construct.Name),
		"UnitsLength": len(construct.Units),
		"UnitsArray": func() string {
			lambdas := []string{}
			for _, unit := range construct.Units {
				lambdas = append(lambdas, fmt.Sprintf("[](chisel::ChiselIstream &reader) -> std::unique_ptr<chisel::Unit> { return std::make_unique<chisel::%s>(reader); }", fetchUnitData(unit).ClassName))
			}
			return "{" + strings.Join(lambdas, ",") + "}"
		}(),
	}

	templ := `
#ifndef {{.Guard}}
#define {{.Guard}}

#include "chisel/chisel.hpp"
#include "units.hpp"

namespace chisel {

class {{.Name}} : public chisel::BufferedConstructReader {
	typedef std::unique_ptr<Unit> (*UnitCreator)(chisel::ChiselIstream &reader);
	static UnitCreator _units[{{.UnitsLength}}];
	int _state;

public:
	{{.Name}}(chisel::ChiselIstream &reader) : chisel::BufferedConstructReader(reader), _state(0) {}
	~{{.Name}}() = default;

	std::unique_ptr<Unit> next() override {
		if (this->_state >= {{.UnitsLength}}) return nullptr;
		return {{.Name}}::_units[this->_state++](this->reader());
	}
};

{{.Name}}::UnitCreator {{.Name}}::_units[{{.UnitsLength}}] = {{.UnitsArray}};

}

#endif // {{.Guard}}
`

	t := template.Must(template.New("").Parse(templ))
	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func generateConstruct(construct *Construct, args Args) error {
	file, err := os.Create(filepath.Join(args.Directory, args.GeneratedConstructDirectory, fmt.Sprintf("%s.hpp", construct.Name)))
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]any{
		"Guard": guard(),
		"Name":  construct.Name,
	}

	templ := `
#ifndef {{.Guard}}
#define {{.Guard}}

#include "chisel/chisel.hpp"

namespace chisel {

class {{.Name}} : public chisel::Construct {
public:
	{{.Name}}(BufferedConstructReader &reader) {
		// Keep constructor signature but do whatever else you want to this class and the changes will persist.
		// Any other constructor will not be used though
	}
};

}

#endif // {{.Guard}}
`

	t := template.Must(template.New("").Parse(templ))
	err = t.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}
