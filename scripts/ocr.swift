#!/usr/bin/swift

import Cocoa
import Vision

// Check for image path argument
guard CommandLine.arguments.count > 1 else {
    print("Usage: ocr.swift <image_path>")
    exit(1)
}

let imagePath = CommandLine.arguments[1]

// Load image
guard let image = NSImage(contentsOfFile: imagePath) else {
    print("Error: Could not load image at \(imagePath)")
    exit(1)
}

// Get CGImage
guard let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
    print("Error: Could not create CGImage")
    exit(1)
}

// Create request handler
let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])

// Store all text observations with their positions
var textBlocks: [(String, CGFloat)] = []

// Create request
let request = VNRecognizeTextRequest { request, error in
    if let error = error {
        fputs("Error: \(error.localizedDescription)\n", stderr)
        return
    }
    
    guard let observations = request.results as? [VNRecognizedTextObservation] else {
        fputs("Error: No text observations\n", stderr)
        return
    }
    
    // Sort observations by Y position (top to bottom)
    // Note: Vision uses bottom-left origin, so we sort by (1 - y) for top-to-bottom
    let sortedObservations = observations.sorted { 
        (1 - $0.boundingBox.origin.y) < (1 - $1.boundingBox.origin.y) 
    }
    
    for observation in sortedObservations {
        if let candidate = observation.topCandidates(1).first {
            let text = candidate.string
            let y = observation.boundingBox.origin.y
            textBlocks.append((text, y))
        }
    }
}

// Configure for accurate recognition
request.recognitionLevel = .accurate
request.recognitionLanguages = ["zh-Hant", "zh-Hans", "en-US"]
request.usesLanguageCorrection = true

// Perform request
do {
    try handler.perform([request])
} catch {
    fputs("Error performing OCR: \(error.localizedDescription)\n", stderr)
    exit(1)
}

// Filter out UI elements and extract main content
let uiPatterns = [
    "Antigravity", "File", "Edit", "Selection", "View", "Go", "Run",
    "Terminal", "Window", "Help", "Extensions", "GitHub", "applejobs",
    "package ", "import (", "import(", "func ", "type ", "const ",
    "//", "/*", "*/", ".go", ".swift", ".py", ".js", ".ts",
    "〉", "›", ">", "internal", "controller", "screenshots"
]

var filteredLines: [String] = []
var currentGroup: [String] = []
var lastY: CGFloat = -1

for (text, y) in textBlocks {
    // Skip very short lines (likely UI fragments)
    let trimmed = text.trimmingCharacters(in: .whitespaces)
    if trimmed.count < 2 {
        continue
    }
    
    // Skip lines matching UI patterns
    var isUI = false
    for pattern in uiPatterns {
        if trimmed.hasPrefix(pattern) || trimmed.contains(pattern) {
            isUI = true
            break
        }
    }
    
    if !isUI {
        // Group lines that are close together vertically
        if lastY >= 0 && abs(y - lastY) > 0.05 {
            // New group - add spacing
            if !currentGroup.isEmpty {
                filteredLines.append(contentsOf: currentGroup)
                filteredLines.append("")
                currentGroup = []
            }
        }
        currentGroup.append(trimmed)
        lastY = y
    }
}

// Add remaining group
if !currentGroup.isEmpty {
    filteredLines.append(contentsOf: currentGroup)
}

// Output filtered text
for line in filteredLines {
    print(line)
}
