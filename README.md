Generates JSON Schema-compatible recipes from images using Open AI's
GPT-4-Turbo-Vision preview API.

You will need an Open AI platform account.
https://platform.openai.com/

To get access to the GPT-4-Turbo-Vision API, you will need to fund your
account with at least 5 dollars in credits. I was able to process roughly 200
recipes (plus 10-15 test calls while I was building this script) for about
nine dollars.

Usage:
1. Create a folder called "out".
2. If generating from a PDF, convert each page to a JPG image. If you're
   using macOS, this is easy to do using [Automator](https://discussions.apple.com/thread/3311405).
3. Remove any images that do not contain a recipe.
4. Place all the images to be converted into a folder next to gpt.go.
5. Set the `input_folder` variable to your image folder name.
6. Set the `author` variable to any value (perhaps the author of the recipes
   you're converting).
7. Set `key` to your OpenAI key.
8. Run the file by invoking `go run ./gpt.go`.

The Vision API is in preview at the time of this writing. Rate limits are
low. Depending on the number of requests, you will hit these limits and start
seeing errors. When you do, just kill the script and try again later. So long
as you don't move files out of `out`, it will pick up where it left off.