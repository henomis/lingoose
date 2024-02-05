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

func NewYoutubeDL() *YoutubeDLLoader {
	return &YoutubeDLLoader{
		youtubeDlPath: defaultYoutubeDLPath,
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

func (y *YoutubeDLLoader) LoadFromSource(ctx context.Context, source string) ([]document.Document, error) {
	y.path = source
	return y.Load(ctx)
}

func (y *YoutubeDLLoader) loadVideo(ctx context.Context) ([]document.Document, error) {
	tempDir, err := os.MkdirTemp("", "youtube-dl")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	args := []string{
		y.subtitlesMode,
		"--sub-lang", y.language,
		"--skip-download",
		"-o", fmt.Sprintf("%s/subtitles", tempDir),
		y.path,
	}

	//nolint:gosec
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
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var plainText string
	for _, line := range lines {
		timestampRegex := regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}`)
		line = timestampRegex.ReplaceAllString(line, "")

		cueSettingsRegex := regexp.MustCompile(`(<c[.\w\s]+>|<\/c>)`)
		line = cueSettingsRegex.ReplaceAllString(line, "")

		vttTagsRegex := regexp.MustCompile(`(<\/?\w+>)`)
		line = vttTagsRegex.ReplaceAllString(line, "")

		line = strings.ReplaceAll(line, "&nbsp;", "")

		line = strings.TrimSpace(line)
		if line != "" {
			plainText += line + "\n"
		}
	}

	if errScanner := scanner.Err(); errScanner != nil {
		return "", errScanner
	}

	return plainText, nil
}
