# IMG-CLI: Advanced Image Generation Tool

A sophisticated command-line tool that uses Google's Gemini API to transform portraits with different outfits and styles. The application provides powerful image analysis, generation, and workflow capabilities with intelligent caching for optimal performance.

## 🚀 Features

### Core Capabilities
- **Outfit Analysis & Generation**: Extract and apply detailed clothing descriptions
- **Style Transfer**: Apply visual/photographic styles from reference images
- **Art Style Transfer**: Apply artistic styles to images or generate from text
- **Hair Preservation**: Maintain or modify hairstyles independently
- **Batch Processing**: Process multiple images with consistent transformations
- **Smart Caching**: Automatic caching of analyses for improved performance

### Analyzers
- **Outfit Analyzer**: Extracts comprehensive outfit details including:
  - Clothing items with fabric, construction, and hardware details
  - Style genre, formality, and aesthetic influences
  - Hair color, style, length, and texture
  - Accessories (excluding glasses)
- **Visual Style Analyzer**: Identifies photographic characteristics:
  - Lighting setup and direction
  - Color grading and mood
  - Composition and framing
  - Background elements
- **Art Style Analyzer**: Analyzes artistic styles for replication

### Generators
- **Outfit Generator**: Generates images with different outfits while preserving facial features
- **Style Transfer Generator**: Applies visual styles to images
- **Combined Generator**: Applies both outfit and style transformations
- **Art Style Generator**: Creates images in specific artistic styles
- **Style Guide Generator**: Creates style reference guides

### Workflows
- **outfit-variations**: Generate multiple outfit variations for a portrait
- **style-transfer**: Apply visual styles from reference images
- **complete-transformation**: Combine outfit and style changes
- **cross-reference**: Mix outfit and style from different sources
- **outfit-swap**: Apply an outfit to specific subjects with style options
- **use-art-style**: Apply artistic styles to images or text prompts
- **analyze-style**: Batch analyze artistic styles in images
- **create-style-guide**: Generate comprehensive style guides

## 📦 Installation

### Prerequisites
- Go 1.23.0 or later
- Gemini API key

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go
   ```

2. **Set up your Gemini API key**

   Create a `.env` file:
   ```env
   GEMINI_API_KEY=your_api_key_here
   ```

   Or set as environment variable:
   ```bash
   export GEMINI_API_KEY=your_api_key_here
   ```

3. **Build the application**
   ```bash
   go build -o img-cli.exe
   ```

   Or run directly without building:
   ```bash
   go run main.go [command]
   ```

## 📂 Directory Structure

```
subjects/            # Input portrait images
  ├── person1.jpg
  └── person2.png

outfits/            # Reference outfit images
  ├── .cache/       # Cached outfit analyses (auto-generated)
  ├── suit.png
  └── casual.jpg

styles/             # Style reference images
  ├── .cache/       # Cached style analyses (auto-generated)
  ├── dramatic.png
  └── soft.jpg

output/             # Generated images (auto-organized)
  └── YYYY-MM-DD/
      └── HHMMSS/
          └── generated_images.png
```

## 💻 Usage

### Basic Commands

#### Analyze Images
```bash
# Analyze all aspects of an image
./img-cli.exe analyze image.jpg

# Analyze specific aspect
./img-cli.exe analyze outfit ./outfits/suit.png
./img-cli.exe analyze visual_style ./styles/dramatic.png
./img-cli.exe analyze art_style ./styles/artwork.jpg

# Skip cache
./img-cli.exe analyze outfit image.jpg --no-cache
```

#### Generate Images
```bash
# Generate with text description
./img-cli.exe generate portrait.jpg "business suit" --type outfit

# Generate with reference image
./img-cli.exe generate portrait.jpg --type outfit --outfit-ref ./outfits/suit.png

# Include original reference in API request for better accuracy
./img-cli.exe generate portrait.jpg --type outfit --outfit-ref ./outfits/suit.png --send-original

# Style transfer
./img-cli.exe generate image.jpg "dramatic lighting" --type style_transfer --style-ref ./styles/dramatic.png
```

### Advanced Workflows

#### Outfit Variations
Generate multiple outfit variations for a portrait:
```bash
./img-cli.exe workflow outfit-variations ./subjects/person.png
```

#### Style Transfer
Apply a visual style to an image:
```bash
./img-cli.exe workflow style-transfer ./subjects/person.jpg --style-ref ./styles/dramatic.png
```

#### Complete Transformation
Change both outfit and style:
```bash
./img-cli.exe workflow complete-transformation ./subjects/person.jpg --outfit-ref ./outfits/suit.png --style-ref ./styles/professional.png
```

#### Cross-Reference
Combine outfit and style from different sources:
```bash
./img-cli.exe workflow cross-reference ./subjects/person.jpg --outfit-ref ./outfits/casual.png --style-ref ./styles/outdoor.png
```

#### Outfit Swap
Apply an outfit to specific subjects with optional style:
```bash
# Apply outfit to specific test subject
./img-cli.exe workflow outfit-swap ./outfits/suit.png --test person1

# Apply outfit with random style from directory
./img-cli.exe workflow outfit-swap ./outfits/suit.png --style-ref ./styles/batch/ --test person1

# Apply outfit to all subjects in directory
./img-cli.exe workflow outfit-swap ./outfits/suit.png
```

#### Art Style Transfer
Apply artistic styles:
```bash
# From text prompt
./img-cli.exe workflow use-art-style "a peaceful forest" --style-ref ./styles/watercolor.png

# From image
./img-cli.exe workflow use-art-style ./subjects/photo.jpg --style-ref ./styles/oil-painting.png
```

### Cache Management

The application automatically caches analysis results for 7 days to improve performance.

```bash
# View cache statistics
./img-cli.exe cache stats

# Clear all cache
./img-cli.exe cache clear

# Clear specific cache type
./img-cli.exe cache clear-outfit
./img-cli.exe cache clear-visual_style
./img-cli.exe cache clear-art_style
```

### Global Options

```bash
# Set custom log level
./img-cli.exe --log-level DEBUG [command]

# Use JSON logging
./img-cli.exe --json-log [command]

# Use custom config file
./img-cli.exe --config custom.env [command]

# Provide API key directly
./img-cli.exe --api-key YOUR_KEY [command]
```

## 🎨 How It Works

### Outfit Generation Process
1. User provides outfit input (text description or image path)
2. If image path provided:
   - Check for cached analysis in `outfits/.cache/`
   - If not cached, analyze the outfit image
   - Extract detailed clothing descriptions
3. Generate new portraits using the outfit description
4. Optional: Include reference image with `--send-original` for more accurate results

### Style Transfer Process
1. Analyze the style reference image for visual characteristics
2. Extract lighting, color grading, composition details
3. Apply these characteristics to the target image
4. Preserve subject identity while transforming style

### Caching System
- Analyses are cached in directory-specific `.cache` folders
- Cache key based on filename (not full path)
- TTL: 7 days by default
- File hash validation ensures cache accuracy

## ⚡ Performance Optimizations

- **Connection Pooling**: Reuses HTTP connections
- **Smart Caching**: Reduces redundant API calls
- **Concurrent Processing**: Worker pools for batch operations
- **Rate Limiting**: Built-in rate limiting (2 req/sec default)
- **Retry Logic**: Exponential backoff for failed requests

## 🏗️ Architecture

The application follows a modular architecture with clear separation of concerns:

```
cmd/                    # CLI commands (cobra-based)
├── root.go            # Root command and global flags
├── analyze.go         # Analysis command
├── generate.go        # Generation command
├── workflow.go        # Workflow command
└── cache.go          # Cache management command

pkg/
├── analyzer/          # Image analysis modules
│   ├── interface.go
│   ├── outfit.go
│   ├── visual.go
│   └── art_style.go
├── generator/         # Image generation modules
│   ├── interface.go
│   ├── outfit.go
│   ├── style.go
│   └── combined.go
├── workflow/          # Workflow orchestration
│   ├── orchestrator.go
│   └── types.go
├── gemini/           # Gemini API client
│   ├── client.go
│   └── types.go
├── cache/            # Caching system
│   ├── cache.go
│   └── optimized.go
├── logger/           # Structured logging
├── errors/           # Custom error types
├── models/           # Data models
├── client/           # Optimized HTTP client
└── concurrent/       # Concurrency utilities
```

## 🔧 Configuration

### Environment Variables
- `GEMINI_API_KEY`: Your Gemini API key (required)

### API Configuration
- Model: `gemini-2.0-flash-exp`
- Timeout: 180 seconds
- Rate limit: 2 requests/second (configurable)

## 📝 Important Notes

### Material Descriptions
The analyzer describes materials as genuine (e.g., "leather" not "faux leather") for fashion design accuracy.

### Prompt Engineering
The system uses detailed prompts optimized for:
- 9:16 portrait format (vertical)
- Waist-up framing
- Pure black background
- Natural pose variation
- Exact facial feature preservation

### Best Practices
- Always use absolute file paths when possible
- Cache is automatically managed, no manual intervention needed
- Use `--send-original` flag for more accurate outfit replication
- Place reference images in appropriate directories for organization

## 🤝 Contributing

Contributions are welcome! Please ensure:
- Code follows existing patterns and conventions
- Tests are included for new features
- Documentation is updated accordingly

## 📄 License

[Your License Here]

## 🆘 Support

For issues or questions, please open an issue on the repository.