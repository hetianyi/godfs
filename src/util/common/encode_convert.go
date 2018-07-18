package common

import "github.com/axgle/mahonia"

func Convert(input string, srcEnc string, tarEnc string) string {
    srcCoder := mahonia.NewDecoder(srcEnc)
    srcResult := srcCoder.ConvertString(input)
    tagCoder := mahonia.NewDecoder(tarEnc)
    _, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
    return string(cdata)
}
