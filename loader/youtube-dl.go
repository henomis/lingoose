package loader

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

var (
	ErrYoutubeDLNotFound             = fmt.Errorf("youtube-dl not found")
	defaultYoutubeDLPath             = "/usr/bin/youtube-dl"
	defaultYoutubeDLSubtitleLanguage = "en"
	defaultYoutubeDLSubtitleMode     = "--write-sub"
)

type YoutubeDLLoader struct {
	loader Loader

	youtubeDlPath string
	path          string
	language      string
	subtitlesMode string
}

func NewYoutubeDLLoader(url string) *YoutubeDLLoader {
	return &YoutubeDLLoader{
		youtubeDlPath: defaultYoutubeDLPath,
		path:          url,
		language:      defaultYoutubeDLSubtitleLanguage,
		subtitlesMode: defaultYoutubeDLSubtitleMode,
	}
}

func (y *YoutubeDLLoader) WithYoutubeDLPath(youtubeDLPath string) *YoutubeDLLoader {
	y.youtubeDlPath = youtubeDLPath
	return y
}

func (y *YoutubeDLLoader) WithTextSplitter(textSplitter TextSplitter) *YoutubeDLLoader {
	y.loader.textSplitter = textSplitter
	return y
}

func (y *YoutubeDLLoader) WithLanguage(language string) *YoutubeDLLoader {
	y.language = language
	return y
}

func (y *YoutubeDLLoader) WithAutoSubtitlesMode() *YoutubeDLLoader {
	y.subtitlesMode = "--write-auto-sub"
	return y
}

func (y *YoutubeDLLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(y.youtubeDlPath)
	if err != nil {
		return nil, ErrYoutubeDLNotFound
	}

	documents, err := y.loadVideo(ctx)
	if err != nil {
		return nil, err
	}

	if y.loader.textSplitter != nil {
		documents = y.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (y *YoutubeDLLoader) loadVideo(ctx context.Context) ([]document.Document, error) {

	// create a temporary directory
	tempDir, err := os.MkdirTemp("", "youtube-dl")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	args := []string{
		y.subtitlesMode,
		"--sub-lang", y.language,
		"--skip-download",
		"--sub-format", "srt",
		"--convert-subs", "srt",
		"-o", fmt.Sprintf("%s/subtitles", tempDir),
		y.path,
	}

	cmd := exec.CommandContext(ctx, y.youtubeDlPath, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	plainText, err := convertVTTtoPlainText(fmt.Sprintf("%s/subtitles.%s.vtt", tempDir, y.language))
	if err != nil {
		return nil, err
	}

	return []document.Document{
		{
			Content: plainText,
			Metadata: types.Meta{
				"source": y.path,
			},
		},
	}, nil
}

func convertVTTtoPlainText(filename string) (string, error) {
	// Read the VTT file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Remove VTT-specific tags and convert to plain text
	var plainText string
	for _, line := range lines {
		// Remove timestamp tags
		timestampRegex := regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}`)
		line = timestampRegex.ReplaceAllString(line, "")

		// Remove cue settings tags
		cueSettingsRegex := regexp.MustCompile(`(<c[.\w\s]+>|<\/c>)`)
		line = cueSettingsRegex.ReplaceAllString(line, "")

		// Remove other VTT tags
		vttTagsRegex := regexp.MustCompile(`(<\/?\w+>)`)
		line = vttTagsRegex.ReplaceAllString(line, "")

		// Remove &nbsp;
		line = strings.ReplaceAll(line, "&nbsp;", "")

		// Trim leading/trailing spaces and append to plain text
		line = strings.TrimSpace(line)
		if line != "" {
			plainText += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return plainText, nil
}
