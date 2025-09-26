# Image Generation Application Documentation

## Overview
This repository contains a sophisticated image generation application that uses Google's Gemini API to transform portraits with different outfits and styles.

## Application Structure

### Advanced Interface (`cmd/main.go`)
- **Purpose**: Full-featured image processing with multiple workflows
- **Commands**:
  - `analyze`: Analyze images for outfit, visual style, or art style
  - `generate`: Generate images with specific transformations
  - `workflow`: Run complex multi-step workflows
  - `batch`: Legacy batch processing
  - `cache`: Manage analysis cache

## Directory Structure
```
subjects/            # Input portrait images

outfits/            # Reference outfit images
  .cache/           # Cached outfit analyses (JSON)
    outfit_*.json   # Detailed outfit descriptions

styles/             # Style reference images
  .cache/           # Cached style analyses (JSON)
    visual_style_*.json
    art_style_*.json

output/             # Generated images (organized by date/time)
  YYYY-MM-DD/
    HHMMSS/
```

## Key Components

### Cache System (`pkg/cache/cache.go`)
- Stores detailed analyses in directory-specific locations:
  - Outfit analyses: `outfits/.cache/`
  - Style analyses: `styles/.cache/`
- Cache entries include:
  - Detailed clothing descriptions
  - Style analysis
  - Colors, accessories, hair details
- TTL: 7 days by default
- Key generation based on filename (not full path)

### Analyzers (`pkg/analyzer/`)
- **OutfitAnalyzer**: Extracts comprehensive outfit details
  - Clothing items with fabric, construction, hardware details
  - Style genre, formality, aesthetic influences
  - Hair color, style, length, texture
  - Accessories (excluding glasses)
- **VisualStyleAnalyzer**: Analyzes visual/photographic style
- **ArtStyleAnalyzer**: Analyzes artistic style

### Workflows (`pkg/workflow/`)
- outfit-variations: Generate multiple outfit variations
- style-transfer: Apply visual styles
- complete-transformation: Full outfit + style change
- cross-reference: Combine multiple references
- outfit-swap: Swap outfits between subjects

## How Outfit Generation Works

1. User provides outfit input (text or image path)
2. If image path:
   - Check for cached analysis in `outfits/.cache/outfit_[filename].json`
   - If found: Use detailed description (clothing list, style, overall aesthetic)
   - If not found: Analyze the image first
3. Generate new portraits using the description
4. With `--send-original` flag: Also include the actual outfit image in the API request for more authentic results

## API Integration
- Uses Gemini API (specifically gemini-2.0-flash-exp model)
- Requires GEMINI_API_KEY environment variable
- Request includes:
  - Portrait image from subjects/
  - Text prompt with outfit description
  - Optionally: Reference outfit image (with --send-original)

## Important Notes

### Material Descriptions
The analyzer is configured to always describe materials as genuine (never "faux"):
- Describes as "leather" not "faux leather"
- Describes as "fur" not "faux fur"
- This is intentional for fashion design accuracy

### Prompt Engineering
The system uses detailed prompts for:
- 9:16 portrait format (vertical)
- Waist-up framing
- Pure black background
- Natural pose variation
- Exact facial feature preservation

### Performance Considerations
- 2-second delay between API requests to avoid rate limiting
- Cache system reduces redundant API calls
- File hashing ensures cache validity

## Common Commands

### CLI Interface (img-cli.exe or go run cmd/main.go)
```bash
# Build the CLI
go build -o img-cli.exe

# Analyze an outfit
./img-cli.exe analyze outfit ./outfits/business-suit.png

# Clear outfit cache
./img-cli.exe cache clear-outfit

# Run outfit variations workflow
./img-cli.exe workflow outfit-variations ./subjects/person.png

# Outfit swap workflow (applies outfit to all subjects)
./img-cli.exe workflow outfit-swap ./outfits/gather/outfit_28.jpg --style-ref ./styles/plain-white.png

# Cross-reference workflow (outfit + style)
./img-cli.exe workflow cross-reference ./subjects/person.jpg --outfit-ref ./outfits/suit.png --style-ref ./styles/dramatic.png
```

## Building the Application

```bash
go build -o img-cli.exe
```

This creates `img-cli.exe` which supports workflows, analysis, caching, and complex image generation operations.

## Recent Updates
- Added `--send-original` flag to include outfit reference images in API requests
- Integrated cache system for improved performance
- Improved outfit description extraction from cached analyses