package typescript

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gokick/tools/codegen"
	"gokick/tools/tsgen/catalog"
	"gokick/tools/tsgen/gosource"
)

const guardsModule = "@/shared/TypeGuards/typeGuards"

const messageModule = "@/shared/Fetch/Envelope/ApiMessage"

const fieldErrorsModule = "@/shared/Fetch/Envelope/ApiFieldErrors"

var typeNames = map[catalog.Kind]string{
	catalog.String: "string",

	catalog.Number: "number",

	catalog.Boolean: "boolean",

	catalog.Record: "Record<string, unknown>",
}

var guardNames = map[catalog.Kind]string{
	catalog.String: "isString",

	catalog.Number: "isNumber",

	catalog.Boolean: "isBoolean",

	catalog.Record: "isRecord",
}

func unionBlock(u *catalog.Union) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export const %s = {\n", u.Dir.Name)

	for _, m := range u.Members {
		fmt.Fprintf(&b, "    %s: %s,\n", m.Name, codegen.Quote(m.Value))
	}

	fmt.Fprintf(&b, "} as const;\n\nexport type %[1]s = (typeof %[1]s)[keyof typeof %[1]s];\n\n", u.Dir.Name)
	fmt.Fprintf(&b, "export const is%[1]s = (v: unknown): v is %[1]s =>\n    Object.values(%[1]s).some((value) => value === v);\n", u.Dir.Name)

	return b.String()
}

func (f *file) addDTO(d *catalog.DTO) error {
	var b strings.Builder
	fmt.Fprintf(&b, "export type %s = {\n", d.Dir.Name)

	for _, fl := range d.Fields {
		tsType, err := f.tsType(fl.Shape)
		if err != nil {
			return fmt.Errorf("%s field %s: %w", d.Ref, fl.Name, err)
		}

		optional := ""
		if fl.Optional {
			optional = "?"
		}

		fmt.Fprintf(&b, "    %s%s: %s;\n", fl.Name, optional, tsType)
	}

	b.WriteString("};\n")

	if d.Dir.NoGuard == false {
		fmt.Fprintf(&b, "\nexport const is%[1]s = (v: unknown): v is %[1]s =>\n    %[2]s(v)", d.Dir.Name, f.helper("isRecord"))

		for _, fl := range d.Fields {
			guard, err := f.guard(fl)
			if err != nil {
				return fmt.Errorf("%s field %s: %w", d.Ref, fl.Name, err)
			}

			fmt.Fprintf(&b, "\n    && %s(v['%s'])", guard, fl.Name)
		}

		b.WriteString(";\n")
	}

	f.blocks = append(f.blocks, b.String())

	if d.Dir.Request == false {
		return nil
	}

	name := d.Dir.Name + "Errors"
	if err := f.claim(name); err != nil {
		return err
	}

	f.blocks = append(f.blocks, f.errorsBlock(name, d.ErrorPaths))

	return nil
}

func (f *file) errorsBlock(name string, paths []string) string {
	f.use(messageModule, "type ApiMessage")
	f.use(fieldErrorsModule, "isApiFieldErrors")

	quote := slices.ContainsFunc(paths, func(path string) bool {
		return strings.Contains(path, catalog.Index) == false && strings.Contains(path, ".")
	})
	keys := make([]string, 0, len(paths))
	var b strings.Builder
	fmt.Fprintf(&b, "export type %s = {\n", name)

	for _, path := range paths {
		switch {
		case strings.Contains(path, catalog.Index):
			fmt.Fprintf(&b, "    [key: `%s`]: ApiMessage;\n", strings.ReplaceAll(path, catalog.Index, "[${number}]"))
		case quote:
			fmt.Fprintf(&b, "    %s?: ApiMessage;\n", codegen.Quote(path))
		default:
			fmt.Fprintf(&b, "    %s?: ApiMessage;\n", path)
		}

		keys = append(keys, strings.ReplaceAll(regexp.QuoteMeta(path), regexp.QuoteMeta(catalog.Index), `\[\d+\]`))
	}

	fmt.Fprintf(&b, "};\n\nexport const is%[1]s = (v: unknown): v is %[1]s =>\n    isApiFieldErrors(/^(?:%[2]s)$/, v);\n", name, strings.Join(keys, "|"))

	return b.String()
}

func (f *file) tsType(s catalog.Shape) (string, error) {
	name := typeNames[s.Kind]
	if s.Kind == catalog.Named {
		dir, err := f.named(s.Ref)
		if err != nil {
			return "", err
		}

		name = dir.Name
		f.use(dir.Module(), "type "+dir.Name)
	}

	if s.Array {
		name += "[]"
	}

	if s.Nullable {
		name += " | null"
	}

	return name, nil
}

func (f *file) guard(fl catalog.Field) (string, error) {
	guard := guardNames[fl.Shape.Kind]
	if fl.Shape.Kind == catalog.Named {
		dir, err := f.named(fl.Shape.Ref)
		if err != nil {
			return "", err
		}

		if dir.NoGuard {
			return "", fmt.Errorf("%s has no guard (noguard or request), so a guarded type cannot contain it", dir.Name)
		}

		guard = "is" + dir.Name
		f.use(dir.Module(), guard)
	} else {
		f.helper(guard)
	}

	if fl.Shape.Array {
		guard = f.helper("arrayOf") + "(" + guard + ")"
	}

	if fl.Shape.Nullable {
		guard = f.helper("nullable") + "(" + guard + ")"
	}

	if fl.Optional {
		guard = f.helper("optional") + "(" + guard + ")"
	}

	return guard, nil
}

func (f *file) named(ref gosource.TypeRef) (catalog.Directive, error) {
	dir, ok := f.directives[ref]
	if ok == false {
		return catalog.Directive{}, fmt.Errorf("%s has no %s directive", ref, catalog.Prefix)
	}

	return dir, nil
}

func (f *file) helper(name string) string {
	f.use(guardsModule, name)

	return name
}
