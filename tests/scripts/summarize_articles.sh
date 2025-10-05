#!/bin/bash

# Script to create summaries for each article one by one

echo "📰 Article Summary Generator"
echo "============================="

# Array of article URLs
articles=(
    "https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"
    "https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/"
    "https://techcrunch.com/2025/07/27/itch-io-is-the-latest-marketplace-to-crack-down-on-adult-games/"
    "https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/"
    "https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/"
    "https://techcrunch.com/2025/07/26/dating-safety-app-tea-breached-exposing-72000-user-images/"
    "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"
    "https://techcrunch.com/2025/07/25/intel-is-spinning-off-its-network-and-edge-group/"
    "https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/"
    "https://techcrunch.com/2025/07/27/doge-has-built-an-ai-tool-to-slash-federal-regulations/"
    "https://edition.cnn.com/2025/07/27/business/us-china-trade-talks-stockholm-intl-hnk"
    "https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal"
    "https://edition.cnn.com/2025/07/27/business/eu-trade-deal"
    "https://edition.cnn.com/2025/07/26/tech/daydream-ai-online-shopping"
    "https://edition.cnn.com/2025/07/25/tech/meta-ai-superintelligence-team-who-its-hiring"
    "https://edition.cnn.com/2025/07/25/tech/sequoia-islamophobia-maguire-mamdani"
    "https://edition.cnn.com/2025/07/24/tech/intel-layoffs-15-percent-q2-earnings"
)

# Create output directory
mkdir -p ../../test_results/summaries

# Counter for article numbering
counter=1

echo "Starting to process ${#articles[@]} articles..."
echo ""

# Loop through each article
for url in "${articles[@]}"; do
    echo "📄 Processing Article $counter/17:"
    echo "URL: $url"
    echo "Generating summary..."
    
    # Create summary using the API
    response=$(curl -s -X POST http://localhost:8080/chat \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"Summarize this article: $url\"}")
    
    # Extract just the answer content
    summary=$(echo "$response" | jq -r '.answer' 2>/dev/null || echo "$response")
    
    # Create filename based on counter
    filename="test_results/summaries/article_$(printf "%02d" $counter)_summary.txt"
    
    # Write summary to file
    cat > "$filename" << EOF
Article $counter Summary
=======================
URL: $url
Generated: $(date)

SUMMARY:
$summary

EOF
    
    echo "✅ Summary saved to: $filename"
    echo ""
    
    # Increment counter
    ((counter++))
    
    # Small delay to avoid overwhelming the API
    sleep 1
done

echo "🎉 All summaries completed!"
echo "📁 Check the 'test_results/summaries' directory for individual summary files."
echo "📊 Total articles processed: $((counter-1))"
