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

// Create request
let request = VNRecognizeTextRequest { request, error in
    if let error = error {
        print("Error: \(error.localizedDescription)")
        return
    }
    
    guard let observations = request.results as? [VNRecognizedTextObservation] else {
        print("Error: No text observations")
        return
    }
    
    // Extract and print text
    for observation in observations {
        if let candidate = observation.topCandidates(1).first {
            print(candidate.string)
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
    print("Error performing OCR: \(error.localizedDescription)")
    exit(1)
}
