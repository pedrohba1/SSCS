# Self-Sovereign Camera System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<a href="https://pkg.go.dev/github.com/pedrohba1/SSCS/services"><img src="https://pkg.go.dev/badge/github.com/pedrohba1/SSCS/services.svg" alt="Go Reference"></a>

The Self-Sovereign Camera System (SSCS) is an open-source, decentralized camera surveillance solution with integrated facial and human detection capabilities.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contribution](#contribution)
- [License](#license)

## Features

1. **Facial Detection**: Capable of detecting faces in real-time.
2. **Human Detection**: Real-time detection of human figures.
3. **Facial Database**: You can save faces for later recognition and notifications.
4. **Multiple Video Feeds**: If you have multiple cameras you can integrate multiple video feeds.  
5. **Archiving and Playback**: integrate either your local storage our cloud storage to have later access to files.
6. **Notifications**: This software can notify you when some specific events happen on camera, in your phone. 
7. **Self-Sovereign**: Your data, your rules. No third-party control or access, unless you prefer it. In that case, it's up to the user to integrate the tools provided.
8. **Search and Filter**: with the saved data, you can check for when specific events happened in the feed. 

## Installation

### Prerequisites
- GoLang (v1.19+ recommended)
- OpenCV (v4.0+ recommended)
- ffmpeg (version 4.4.2 recommended)

To quickly install openCV, just clone this repo and run `make install-opencv`. It will attempt to install all dependencies required and setup openCV properly.

## Contribution
Fork the project.
Create a new branch (git checkout -b feature/YourFeature).
Commit your changes (git commit -am 'Add some feature').
Push to the branch (git push origin feature/YourFeature).
Open a pull request.
We welcome contributions! Please read our CONTRIBUTING.md for more information on how to contribute to SSCS.

License
This project is licensed under the MIT License - see the LICENSE.md file for details.