package main

// Generates JSON Schema-compatible recipes from images using Open AI's
// GPT-4-Turbo-Vision preview API.
//
// You will need an Open AI platform account.
// https://platform.openai.com/
//
// To get access to the GPT-4-Turbo-Vision API, you will need to fund your
// account with at least 5 dollars in credits. I was able to process roughly 200
// recipes (plus 10-15 test calls while I was building this script) for about
// nine dollars.
//
// Usage:
// 1. Create a folder called "out".
// 2. If generating from a PDF, convert each page to a JPG image. If you're
//    using macOS, this is easy to do using [Automator](https://discussions.apple.com/thread/3311405).
// 3. Remove any images that do not contain a recipe.
// 4. Place all the images to be converted into a folder next to gpt.go.
// 5. Set the `input_folder` variable to your image folder name.
// 6. Set the `author` variable to any value (perhaps the author of the recipes
//    you're converting).
// 7. Set `key` to your OpenAI key.
// 8. Run the file by invoking `go run ./gpt.go`.
//
// The Vision API is in preview at the time of this writing. Rate limits are
// low. Depending on the number of requests, you will hit these limits and start
// seeing errors. When you do, just kill the script and try again later. So long
// as you don't move files out of `out`, it will pick up where it left off.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	input_folder = ""
	author       = ""
	key          = ""
	prompt       = "Return raw json. do not return anything but json. Do not add markdown " +
		"fences or whitespace. Process this image as a recipe, and convert to a " +
		"JSON document conforming to the schema found at schema.org/recipe. Take " +
		"the title from the image. Write a one sentence creative description. " +
		"Attribute authorship to '" + author + "'. Extract the number of servings " +
		"from the document, then calculate nutrition facts based off ingredients, " +
		"and add it to the final output. Do not explain the output, or add " +
		"supporting documentation. only show the final result. Make sure there " +
		"are no duplicate keys in the output. If there are any 'additional notes', " +
		"append them to the recipeInstructions list. If any extracted text is " +
		"in all caps, correct it to normal sentence casing. Convert metric " +
		"weight measures to imperial."
)

func main() {
	files, err := filepath.Glob(fmt.Sprintf("./%s/*", input_folder))
	if err != nil {
		log.Fatal(err)
	}

	j, err := filepath.Glob("./out/*.json")
	if err != nil {
		log.Fatal(err)
	}

	exists := make(map[string]bool)
	for _, f := range j {
		exists[input_folder+"/"+strings.TrimPrefix(strings.TrimSuffix(f, ".json"), "out/")] = true
	}

	for _, f := range files {
		if _, done := exists[f]; done {
			continue
		}

		b, err := getRecipeJSON(f)
		if err != nil {
			log.Printf("[ERROR] %s: error: %s\n", f, err.Error())
			continue
		}

		log.Printf("[OK] %s\n", f)
		os.WriteFile("out/"+strings.TrimPrefix(f, input_folder+"/")+".json", b, 0777)
		time.Sleep(5 * time.Second)
	}
}

func encodeImage(path string) ([]byte, error) {
	han, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer han.Close()

	b, err := io.ReadAll(han)
	if err != nil {
		return nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

func getRecipeJSON(path string) ([]byte, error) {
	b, err := encodeImage(path)
	if err != nil {
		return nil, err
	}

	payload := NewPayload(b)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("%d: %s", res.StatusCode, string(b))
	}

	var resp Response
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no messages returned")
	}

	return []byte(resp.Choices[0].Message.Content), nil
}

type Response struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Content string `json:"content"`
}

type Payload struct {
	Model     string    `json:"model,omitempty"`
	Messages  []Message `json:"messages,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type Message struct {
	Role    string    `json:"role,omitempty"`
	Content []Content `json:"content,omitempty"`
}

type Content struct {
	Type     string `json:"type,omitempty"`
	Text     string `json:"text,omitempty"`
	ImageURL Image  `json:"image_url,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}

func NewPayload(b []byte) Payload {
	return Payload{
		Model:     "gpt-4-vision-preview",
		MaxTokens: 1200,
		// ResponseFormat: "json_object",
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: Image{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", string(b)),
						},
					},
				},
			},
		},
	}
}
