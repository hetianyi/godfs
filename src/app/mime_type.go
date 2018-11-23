package app

import (
	"container/list"
	"strings"
)

// mime types for browser download header 'Content-Type'
var (
	mimeTypes                    = make(map[string]string)
	webMimeTypes                 = make(map[string]string)
	accessControllerAllowOrigins = list.New()
	enable                       = false
	defaultMimeType              = "application/octet-stream"
)

// if enable it will auto add download header 'Content-Type' before download,
// if it is disabled(by default), download header 'Content-Type' will set to 'application/octet-stream'.
func SetMimeTypesEnable() {

	enable = true

	mimeTypes["html"] = "text/html"
	mimeTypes["shtml"] = "text/html"
	mimeTypes["css"] = "text/css"
	mimeTypes["xml"] = "text/xml"
	mimeTypes["gif"] = "image/gif"
	mimeTypes["jpeg"] = "image/jpeg"
	mimeTypes["jpg"] = "image/jpeg"
	mimeTypes["js"] = "application/javascript"
	mimeTypes["atom"] = "application/atom+xml"
	mimeTypes["rss"] = "application/rss+xml"

	mimeTypes["mml"] = "text/mathml"
	mimeTypes["txt"] = "text/plain"
	mimeTypes["jad"] = "text/vnd.sun.j2me.app-descriptor"
	mimeTypes["wml"] = "text/vnd.wap.wml"
	mimeTypes["htc"] = "text/x-component"

	mimeTypes["png"] = "image/png"
	mimeTypes["tif"] = "image/tiff"
	mimeTypes["tiff"] = "image/tiff"
	mimeTypes["wbmp"] = "image/vnd.wap.wbmp"
	mimeTypes["ico"] = "image/x-icon"
	mimeTypes["jng"] = "image/x-jng"
	mimeTypes["bmp"] = "image/x-ms-bmp"
	mimeTypes["svg"] = "image/svg+xml"
	mimeTypes["svgz"] = "image/svg+xml"
	mimeTypes["webp"] = "image/webp"

	mimeTypes["woff"] = "application/font-woff"
	mimeTypes["jar"] = "application/java-archive"
	mimeTypes["war"] = "application/java-archive"
	mimeTypes["ear"] = "application/java-archive"
	mimeTypes["json"] = "application/json"
	mimeTypes["hqx"] = "application/mac-binhex40"
	mimeTypes["doc"] = "application/msword"
	mimeTypes["pdf"] = "application/pdf"
	mimeTypes["ps"] = "application/postscript"
	mimeTypes["eps"] = "application/postscript"
	mimeTypes["ai"] = "application/postscript"
	mimeTypes["rtf"] = "application/rtf"
	mimeTypes["m3u8"] = "application/vnd.apple.mpegurl"
	mimeTypes["xls"] = "application/vnd.ms-excel"
	mimeTypes["eot"] = "application/vnd.ms-fontobject"
	mimeTypes["ppt"] = "application/vnd.ms-powerpoint"
	mimeTypes["wmlc"] = "application/vnd.wap.wmlc"
	mimeTypes["kml"] = "application/vnd.google-earth.kml+xml"
	mimeTypes["kmz"] = "application/vnd.google-earth.kmz"
	mimeTypes["7z"] = "application/x-7z-compressed"
	mimeTypes["cco"] = "application/x-cocoa"
	mimeTypes["jardiff"] = "application/x-java-archive-diff"
	mimeTypes["jnlp"] = "application/x-java-jnlp-file"
	mimeTypes["run"] = "application/x-makeself"
	mimeTypes["pl"] = "application/x-perl"
	mimeTypes["pm"] = "application/x-perl"
	mimeTypes["prc"] = "application/x-pilot"
	mimeTypes["pdb"] = "application/x-pilot"
	mimeTypes["rar"] = "application/x-rar-compressed"
	mimeTypes["rpm"] = "application/x-redhat-package-manager"
	mimeTypes["sea"] = "application/x-sea"
	mimeTypes["swf"] = "application/x-shockwave-flash"
	mimeTypes["sit"] = "application/x-stuffit"
	mimeTypes["tcl"] = "application/x-tcl"
	mimeTypes["tk"] = "application/x-tcl"
	mimeTypes["der"] = "application/x-x509-ca-cert"
	mimeTypes["pem"] = "application/x-x509-ca-cert"
	mimeTypes["crt"] = "application/x-x509-ca-cert"
	mimeTypes["xpi"] = "application/x-xpinstall"
	mimeTypes["xhtml"] = "application/xhtml+xml"
	mimeTypes["xspf"] = "application/xspf+xml"
	mimeTypes["zip"] = "application/zip"

	mimeTypes["docx"] = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	mimeTypes["xlsx"] = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mimeTypes["pptx"] = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	mimeTypes["mid"] = "audio/midi"
	mimeTypes["midi"] = "audio/midi"
	mimeTypes["kar"] = "audio/midi"
	mimeTypes["mp3"] = "audio/mpeg"
	mimeTypes["ogg"] = "audio/ogg"
	mimeTypes["m4a"] = "audio/x-m4a"
	mimeTypes["ra"] = "audio/x-realaudio"

	mimeTypes["3gpp"] = "video/3gpp"
	mimeTypes["3gp"] = "video/3gpp"
	mimeTypes["ts"] = "video/mp2t"
	mimeTypes["mp4"] = "video/mp4"
	mimeTypes["mpeg"] = "video/mpeg"
	mimeTypes["mpg"] = "video/mpeg"
	mimeTypes["mov"] = "video/quicktime"
	mimeTypes["webm"] = "video/webm"
	mimeTypes["flv"] = "video/x-flv"
	mimeTypes["m4v"] = "video/x-m4v"
	mimeTypes["mng"] = "video/x-mng"
	mimeTypes["asx"] = "video/x-ms-asf"
	mimeTypes["asf"] = "video/x-ms-asf"
	mimeTypes["wmv"] = "video/x-ms-wmv"
	mimeTypes["avi"] = "video/x-msvideo"
}

// get mime type by file ext
func GetContentTypeHeader(fileFormat string) *string {
	if enable {
		format := mimeTypes[strings.TrimLeft(fileFormat, ".")]
		if format == "" {
			return &defaultMimeType
		}
		return &format
	}
	return &defaultMimeType
}

// add web content file format based on mimeTypes
func AddWebMimeType(format string) {
	webMimeTypes[format] = mimeTypes[format]
}

// add web content file format based on mimeTypes
func AddAccessAllowOrigin(origin string) {
	for ele := accessControllerAllowOrigins.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(string) == origin {
			return
		}
	}
	accessControllerAllowOrigins.PushBack(origin)
}

// check whether access origin is allowed
func CheckOriginAllow(origin string) bool {
	if accessControllerAllowOrigins.Len() == 0 {
		return true
	}
	for ele := accessControllerAllowOrigins.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(string) == origin {
			return true
		}
	}
	return false
}

func SupportWebContent(ext string) bool {
	ret := webMimeTypes[strings.TrimLeft(ext, ".")]
	if ret == "" {
		return false
	}
	return true
}
