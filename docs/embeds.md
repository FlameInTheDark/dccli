# Embeds

This CLI supports sending rich messages known as "Embeds". Embeds are defined using JSON.
You can provide the JSON string directly using `--embed` or load it from a file using `--embed-file`.

## JSON Structure

The JSON structure generally follows the [Discord API Embed Object](https://discord.com/developers/docs/resources/channel#embed-object).

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | title of embed |
| `type` | string | type of embed (always "rich" for webhook embeds) |
| `description` | string | description of embed |
| `url` | string | url of embed |
| `timestamp` | ISO8601 timestamp | timestamp of embed content |
| `color` | integer | color code of the embed |
| `footer` | object | footer information |
| `image` | object | image information |
| `thumbnail` | object | thumbnail information |
| `video` | object | video information |
| `provider` | object | provider information |
| `author` | object | author information |
| `fields` | array of objects | fields information |

### Footer Object

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | footer text |
| `icon_url` | string | url of footer icon |

### Image/Thumbnail/Video Object

| Field | Type | Description |
|-------|------|-------------|
| `url` | string | source url of image (only supports http(s) and attachments) |
| `proxy_url` | string | a proxied url of the image |
| `height` | integer | height of image |
| `width` | integer | width of image |

### Author Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | name of author |
| `url` | string | url of author |
| `icon_url` | string | url of author icon |

### Field Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | name of the field |
| `value` | string | value of the field |
| `inline` | boolean | whether or not this field should display inline |

## Examples

### Basic Embed

```json
{
  "title": "Hello World",
  "description": "This is a simple embed.",
  "color": 5814783
}
```

### Complex Embed

```json
{
  "title": "Rich Embed",
  "description": "This embed has **formatting**, fields, and images.",
  "url": "https://discord.com",
  "color": 16711680,
  "timestamp": "2023-10-01T12:00:00.000Z",
  "footer": {
    "text": "Sent via dccli",
    "icon_url": "https://i.imgur.com/fKL31aD.jpg"
  },
  "thumbnail": {
    "url": "https://i.imgur.com/fKL31aD.jpg"
  },
  "image": {
    "url": "https://i.imgur.com/fKL31aD.jpg"
  },
  "author": {
    "name": "Bot Author",
    "url": "https://github.com/FlameInTheDark/dccli",
    "icon_url": "https://i.imgur.com/fKL31aD.jpg"
  },
  "fields": [
    {
      "name": "Field 1",
      "value": "Some value here",
      "inline": true
    },
    {
      "name": "Field 2",
      "value": "Another value",
      "inline": true
    }
  ]
}
```

## Validation

You can validate your embed JSON before sending it using the `validate-embed` command.

```bash
# Validate from string (Bash/zsh/PowerShell with escaping)
dccli messages validate-embed --embed '{"title": "Test"}'
# Note for Windows users: Single quotes might not work as expected in Command Prompt or PowerShell.
# It is recommended to use --embed-file or escape quotes:
# PowerShell: dccli messages validate-embed --embed '{\"title\": \"Test\"}'
# CMD: dccli messages validate-embed --embed "{\"title\": \"Test\"}"

# Validate from file
dccli messages validate-embed --embed-file ./my-embed.json
```
