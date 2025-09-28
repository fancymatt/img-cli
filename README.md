# IMG-CLI: Advanced Image Generation Tool

A sophisticated command-line tool that uses Google's Gemini API to transform portraits with different outfits and styles. The application provides powerful image analysis, generation, and workflow capabilities with intelligent caching for optimal performance.

## 🚀 Features

### Core Capabilities
- **Modular Component System**: Independent control of outfit, hair style/color, makeup, expression, and accessories
- **Outfit Analysis & Generation**: Extract and apply detailed clothing descriptions
- **Style Transfer**: Apply visual/photographic styles from reference images
- **Art Style Transfer**: Apply artistic styles to images or generate from text
- **Hair Preservation**: Separate control of hair style and color
- **Facial Structure Preservation**: Makeup applied as surface layer only
- **Batch Processing**: Process multiple images with all combinations
- **Smart Caching**: Automatic caching of analyses for improved performance

### Analyzers
- **Modular Outfit Analyzer**: Extracts outfit with configurable exclusions
  - Clothing items with fabric, construction, and hardware details
  - Style genre, formality, and aesthetic influences
  - Optionally excludes hair, makeup, or accessories when specified separately
- **Hair Style Analyzer**: Analyzes hairstyle structure only (not color)
  - Cut, shape, and styling techniques
  - Length, texture, volume, and layers
  - Parting and front styling details
- **Hair Color Analyzer**: Analyzes hair color only (not style)
  - Base color, undertones, highlights, and lowlights
  - Coloring techniques (balayage, ombre, etc.)
  - Dimension and special effects
- **Makeup Analyzer**: Analyzes cosmetic application
  - Complexion (foundation, blush, highlighter, contour)
  - Eye makeup (shadow, liner, mascara, brows)
  - Lip color and finish
- **Expression Analyzer**: Analyzes facial expressions
  - Primary emotion and intensity
  - Facial feature positions
  - Gaze direction and mood
- **Accessories Analyzer**: Extracts accessory details
  - Jewelry (earrings, necklaces, bracelets, rings)
  - Bags, belts, scarves, hats, watches
  - Overall accessory styling
- **Visual Style Analyzer**: Identifies photographic characteristics
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
  ├── cache/        # Cached outfit analyses (auto-generated)
  ├── suit.png
  └── casual.jpg

styles/             # Visual/photographic style references
  ├── cache/        # Cached style analyses (auto-generated)
  ├── dramatic.png
  └── soft.jpg

hair-style/         # Hair style references (structure/cut only)
  ├── cache/        # Cached hair style analyses
  ├── professional.png
  └── wavy.jpg

hair-color/         # Hair color references
  ├── cache/        # Cached hair color analyses
  ├── blonde.png
  └── auburn.jpg

makeup/             # Makeup style references
  ├── cache/        # Cached makeup analyses
  ├── natural.png
  └── glamorous.jpg

expressions/        # Facial expression references
  ├── cache/        # Cached expression analyses
  ├── confident.png
  └── serene.jpg

accessories/        # Accessory references
  ├── cache/        # Cached accessory analyses
  ├── jewelry.png
  └── watches.jpg

output/             # Generated images (auto-organized)
  └── YYYY-MM-DD/   # Date folder
      └── HHMMSS/   # Timestamp folder
          ├── outfit_style_subject_timestamp.png
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

#### Outfit Swap (Most Powerful Workflow)

The outfit-swap workflow is the most comprehensive and flexible workflow, supporting modular component control and batch processing.

**Quick Flag Reference:**
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `[outfit]` | - | Outfit image/directory (positional) | `./outfits/shearling-black.png` |
| `--test` | `-t` | Test subjects (omit for all) | All subjects / "jaimee" if -t "" |
| `--style` | `-s` | Photographic style | `./styles/plain-white.png` |
| `--hair-style` | - | Hair style (cut/shape only) | - |
| `--hair-color` | - | Hair color only | - |
| `--makeup` | - | Makeup style | - |
| `--expression` | - | Facial expression | - |
| `--accessories` | `-a` | Accessories (also --accessory) | - |
| `--variations` | `-v` | Variations per combo | 1 |
| `--send-original` | - | Include refs in API | false |
| `--no-confirm` | - | Skip cost prompt | false |
| `--debug` | - | Show debug info | false |

**Basic Usage:**
```bash
# Simple outfit swap (uses all subjects by default)
./img-cli.exe outfit-swap ./outfits/suit.png

# Specify particular subjects
./img-cli.exe outfit-swap ./outfits/suit.png -t "jaimee kat izzy"

# Use default subject (jaimee) with empty -t flag
./img-cli.exe outfit-swap ./outfits/suit.png -t ""
```

**Subject Selection:**
- **No `-t` flag**: Processes ALL subjects in `subjects/` directory
- **`-t ""`**: Uses default subject "jaimee"
- **`-t "name1 name2"`**: Uses specified subjects (without file extensions)

**Style Control:**
```bash
# Add a photographic style
./img-cli.exe outfit-swap ./outfits/suit.png --style ./styles/dramatic.png

# Use default style (plain-white)
./img-cli.exe outfit-swap ./outfits/suit.png -s ./styles/plain-white.png
```

**Modular Component Control:**

The outfit-swap workflow supports independent control of each visual component:

```bash
# Full modular control - Japanese theme example
./img-cli.exe outfit-swap ./outfits/kimono.png \
  --style ./styles/japan.png \
  --hair-style ./hair-style/geisha.png \
  --hair-color ./hair-color/black.png \
  --makeup ./makeup/geisha.png \
  --expression ./expressions/serene.png \
  -a ./accessories/parasol.png \
  -t "jaimee"

# Mix and match components
./img-cli.exe outfit-swap ./outfits/business-suit.png \
  --hair-style ./hair-style/professional.png \
  --makeup ./makeup/natural.png \
  -t "kat izzy"

# Hair style without changing color (preserves subject's natural hair color)
./img-cli.exe outfit-swap ./outfits/casual.png \
  --hair-style ./hair-style/wavy.png

# Hair color without changing style (preserves subject's natural hair style)
./img-cli.exe outfit-swap ./outfits/dress.png \
  --hair-color ./hair-color/blonde.png
```

**Directory Processing (Batch Mode):**

Any component parameter can accept either a single file or a directory. When directories are provided, the workflow creates all possible combinations:

```bash
# Process all outfits with all hair styles for specific subjects
./img-cli.exe outfit-swap ./outfits/ \
  --hair-style ./hair-style/ \
  -t "jaimee kat"

# All combinations: 3 outfits × 2 styles × 2 subjects = 12 images
./img-cli.exe outfit-swap ./outfits/batch/ \
  --style ./styles/batch/ \
  -t "jaimee kat"

# Mix single files and directories
./img-cli.exe outfit-swap ./outfits/ \
  --style ./styles/professional.png \
  --makeup ./makeup/ \
  -t "izzy"
```

**Component Independence:**

Each component is applied independently without affecting others:
- **Outfit**: Applied without including hair, makeup, or accessories if specified separately
- **Hair Style**: Applied without changing hair color (unless --hair-color is also specified)
- **Hair Color**: Applied without changing hair style (unless --hair-style is also specified)
- **Makeup**: Applied as surface layer only, preserving facial structure
- **Expression**: Changes facial expression without altering identity
- **Accessories**: Added without affecting outfit analysis

**Advanced Options:**
```bash
# Generate multiple variations per combination
./img-cli.exe outfit-swap ./outfits/suit.png -v 3

# Include reference images in API request for more accuracy
./img-cli.exe outfit-swap ./outfits/detailed.png --send-original

# Skip cost confirmation prompt
./img-cli.exe outfit-swap ./outfits/ --no-confirm

# Show debug information including prompts
./img-cli.exe outfit-swap ./outfits/test.png --debug
```

**Output Organization:**

Generated images are automatically organized:
```
output/
  └── 2024-01-15/          # Date folder
      └── 143022/          # Timestamp folder
          ├── suit_dramatic_jaimee_20240115_143025.png
          ├── suit_dramatic_kat_20240115_143028.png
          └── suit_dramatic_izzy_20240115_143031.png
```

File naming convention: `{outfit}_{style}_{subject}_{timestamp}.png`

**Complete Example Workflow:**

```bash
# Professional headshots with consistent style
./img-cli.exe outfit-swap ./outfits/business-suit.png \
  --style ./styles/corporate.png \
  --hair-style ./hair-style/professional.png \
  --makeup ./makeup/professional.png \
  --expression ./expressions/confident.png \
  -t "jaimee kat izzy sarah" \
  -v 2 \
  --send-original

# This generates 8 images (4 subjects × 2 variations) with:
# - Business suit outfit
# - Corporate photographic style
# - Professional hairstyle (preserving natural color)
# - Professional makeup (preserving facial structure)
# - Confident expression
# - Original outfit reference included for accuracy
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
./img-cli.exe cache clear-hair_style
./img-cli.exe cache clear-hair_color
./img-cli.exe cache clear-makeup
./img-cli.exe cache clear-expression
./img-cli.exe cache clear-accessories
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

## 🔍 Troubleshooting

### Outfit-Swap Workflow

**Hair color changes when only style specified:**
- The system explicitly preserves original hair color when only `--hair-style` is used
- Ensure you're not also specifying `--hair-color`

**Makeup changes facial structure:**
- Makeup is applied as surface layer only
- The system includes explicit instructions to preserve bone structure
- Report persistent issues with specific makeup references

**Accessories not appearing:**
- Check that accessories are visible in the reference image
- Some items (glasses, weapons) are explicitly excluded
- Verify the cache is up to date with `cache clear-accessories`

**Wrong subjects being processed:**
- No `-t` flag: processes ALL subjects
- `-t ""`: uses only "jaimee"
- `-t "name1 name2"`: uses specified subjects
- Check subjects directory for available files

**Output organization:**
- All images from one command go to the same timestamp folder
- Filename format: `{outfit}_{style}_{subject}_{timestamp}.png`
- Check `output/YYYY-MM-DD/HHMMSS/` for generated images

## 🆘 Support

For issues or questions, please open an issue on the repository.