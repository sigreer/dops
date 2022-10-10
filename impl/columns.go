package impl

import (
	"better-docker-ps/cli"
	"better-docker-ps/docker"
	"better-docker-ps/langext"
	"better-docker-ps/langext/term"
	"better-docker-ps/printer"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ColumnMap = map[string]printer.ColFun{
	"ID":            ColContainerID,
	"Image":         ColFullImage,
	"ImageName":     ColImage,
	"ImageTag":      ColImageTag,
	"Registry":      ColRegistry,
	"ImageRegistry": ColRegistry,
	"Tag":           ColImageTag,
	"Command":       ColCommand,
	"ShortCommand":  ColShortCommand,
	"CreatedAt":     ColCreatedAt,
	"RunningFor":    ColRunningFor,
	"Ports":         ColPorts,
	"State":         ColState,
	"Status":        ColStatus,
	"Size":          ColSize,
	"Names":         ColName,
	"Labels":        ColLabels,
	"LabelKeys":     ColLabelKeys,
	"Mounts":        ColMounts,
	"Networks":      ColNetworks,
	"IP":            ColIP,
}

func ColContainerID(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"CONTAINER ID"}
	}

	if ctx.Opt.Truncate {
		return []string{cont.ID[0:12]}
	} else {
		return []string{cont.ID}
	}
}

func ColFullImage(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"IMAGE"}
	}

	return []string{cont.Image}
}

func ColRegistry(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"REGISTRY"}
	}

	v, _, _ := docker.SplitDockerImage(cont.Image)

	return []string{v}
}

func ColImage(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"IMAGE"}
	}

	_, v, _ := docker.SplitDockerImage(cont.Image)

	return []string{v}
}

func ColImageTag(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"TAG"}
	}

	_, _, v := docker.SplitDockerImage(cont.Image)

	return []string{v}
}

func ColCommand(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"COMMAND"}
	}

	cmd := cont.Command
	if ctx.Opt.Truncate && len(cmd) > 20 {
		cmd = cmd[:19] + "…"
	}

	cmd = "\"" + cmd + "\""

	return []string{cmd}
}

func ColShortCommand(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"COMMAND"}
	}

	spl := strings.Split(cont.Command, " ")
	if len(spl) == 0 {
		return []string{""}
	} else {
		return []string{spl[0]}
	}

}

func ColRunningFor(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"CREATED"}
	}

	ts := time.Unix(cont.Created, 0)
	diff := time.Now().Sub(ts)

	return []string{langext.FormatNaturalDurationEnglish(diff)}
}

func ColCreatedAt(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"CREATED AT"}
	}

	ts := time.Unix(cont.Created, 0)

	return []string{ts.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat)}
}

func ColState(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"STATE"}
	}

	strstate := "[" + strings.ToUpper(string(cont.State)) + "]"

	if !ctx.Opt.OutputColor {
		return []string{strstate}
	}

	return []string{stateColor(cont.State, strstate)}
}

func ColStatus(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"STATUS"}
	}

	if !ctx.Opt.OutputColor {
		return []string{cont.Status}
	}

	return []string{stateColor(cont.State, cont.Status)}
}

func ColPorts(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"PORTS"}
	}

	r := make(map[string]bool)
	for _, port := range cont.Ports {
		p1 := langext.StrPadLeft(strconv.Itoa(port.PublicPort), " ", 5)
		p2 := langext.StrPadLeft(strconv.Itoa(port.PrivatePort), " ", 5)

		if port.PublicPort == 0 {
			r[fmt.Sprintf("         %s / %s", p2, port.Type)] = true
		} else {
			r[fmt.Sprintf("%s -> %s / %s", p1, p2, port.Type)] = true
		}
	}

	return langext.MapKeyArr(r)
}

func ColName(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"NAME"}
	}

	r := make([]string, 0, len(cont.Names))
	for _, n := range cont.Names {
		if len(n) > 0 && n[0] == '/' {
			n = n[1:]
		}
		r = append(r, n)
	}

	return r
}

func ColSize(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"SIZE"}
	}

	if cont.SizeRw == 0 && cont.SizeRootFs == 0 {
		return []string{}
	}

	return []string{fmt.Sprintf("%v (virt %v)", langext.StrPadRight(langext.FormatBytes(cont.SizeRw), " ", 11), langext.FormatBytes(cont.SizeRootFs))}
}

func ColMounts(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"MOUNTS"}
	}

	r := make([]string, 0, len(cont.Mounts))
	for _, mnt := range cont.Mounts {
		val := fmt.Sprintf("%s -> %s", mnt.Source, mnt.Destination)
		if !mnt.RW {
			val += " [ro]"
		}
		r = append(r, val)
	}

	return r
}

func ColIP(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"IP"}
	}

	r := make([]string, 0, len(cont.NetworkSettings.Networks))
	for _, nw := range cont.NetworkSettings.Networks {
		if nw.IPAddress != "" {
			r = append(r, nw.IPAddress)
		}
	}

	return r
}

func ColLabels(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"LABELS"}
	}

	r := make([]string, 0, len(cont.Mounts))
	for k, v := range cont.Labels {
		r = append(r, fmt.Sprintf("%s := %s", k, v))
	}

	return r
}

func ColLabelKeys(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"LABELS"}
	}

	r := make([]string, 0, len(cont.Mounts))
	for k, _ := range cont.Labels {
		r = append(r, k)
	}

	return r
}

func ColNetworks(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
	if cont == nil {
		return []string{"NETWORKS"}
	}

	r := make([]string, 0, len(cont.Mounts))
	for k := range cont.NetworkSettings.Networks {
		r = append(r, k)
	}

	return r
}

func ColPlaintext(str string) printer.ColFun {
	return func(ctx *cli.PSContext, cont *docker.ContainerSchema) []string {
		return []string{str}
	}
}

func getColFun(colkey string) (printer.ColFun, bool) {
	for k, v := range ColumnMap {
		if "{{."+k+"}}" == colkey {
			return v, true
		}
	}
	return nil, false
}

func replaceSingleLineColumnData(ctx *cli.PSContext, data docker.ContainerSchema, format string) string {

	r := format

	for k, v := range ColumnMap {
		r = strings.ReplaceAll(r, "{{."+k+"}}", strings.Join(v(ctx, &data), " "))
	}

	return r

}

func parseTableDef(fmt string) []printer.ColFun {
	split := strings.Split(fmt[6:], "\\t")
	columns1 := make([]printer.ColFun, 0)
	for _, v := range split {
		if cf, ok := getColFun(v); ok {
			columns1 = append(columns1, cf)
		} else {
			columns1 = append(columns1, ColPlaintext(v))
		}
	}
	return columns1
}

func stateColor(state docker.ContainerState, value string) string {
	switch state {
	case docker.StateCreated:
		return term.Yellow(value)
	case docker.StateRunning:
		return term.Green(value)
	case docker.StateRestarting:
		return term.Yellow(value)
	case docker.StateExited:
		return term.Red(value)
	case docker.StatePaused:
		return term.Yellow(value)
	case docker.StateDead:
		return term.Red(value)
	}
	return value
}
