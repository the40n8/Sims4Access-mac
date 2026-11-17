// TS4 Access macOS speech companion.
//
// Drop-in replacement for the Windows SAPI child process that TS4 Access
// spawns from inside the game. Same stdin protocol, one command per line:
//   I<url-encoded text>   stop current speech, then speak text
//   Q<url-encoded text>   speak text after whatever is already queued
//   I                     stop speaking (used by the mod's stop key)
// The process exits when stdin closes, so it never outlives the game.
//
// Options: --voice <name or identifier substring>  --lang <bcp47, e.g. fr-FR>
//          --rate <-10..10, SAPI scale>  --volume <0..100>  --list-voices
//
// MIT licence, like TS4 Access.
import AVFoundation
import Foundation

final class Speaker: NSObject {
    let synth = AVSpeechSynthesizer()
    var voice: AVSpeechSynthesisVoice?
    var rate: Float = AVSpeechUtteranceDefaultSpeechRate
    var volume: Float = 0.8

    func handle(_ line: String) {
        guard let kind = line.first else { return }
        let body = String(line.dropFirst())
        if kind != "Q" {
            synth.stopSpeaking(at: .immediate)
        }
        guard !body.isEmpty else { return }
        let text = body.removingPercentEncoding ?? body
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = voice
        utterance.rate = rate
        utterance.volume = volume
        synth.speak(utterance)
    }
}

func option(_ name: String) -> String? {
    let args = CommandLine.arguments
    guard let i = args.firstIndex(of: name), i + 1 < args.count else { return nil }
    return args[i + 1]
}

let voices = AVSpeechSynthesisVoice.speechVoices()

if CommandLine.arguments.contains("--list-voices") {
    for v in voices.sorted(by: { ($0.language, $0.name) < ($1.language, $1.name) }) {
        print("\(v.language)\t\(v.name)\t\(v.identifier)")
    }
    exit(0)
}

let speaker = Speaker()

let lang = option("--lang")
if let want = option("--voice")?.lowercased(), !want.isEmpty {
    // Prefer a match in the requested language, then any match.
    let matches = voices.filter {
        $0.name.lowercased().contains(want) || $0.identifier.lowercased().contains(want)
    }
    speaker.voice = matches.first(where: { lang != nil && $0.language.hasPrefix(lang!) })
        ?? matches.first
}
if speaker.voice == nil, let lang = lang {
    speaker.voice = AVSpeechSynthesisVoice(language: lang)
}
if let r = option("--rate").flatMap(Float.init) {
    // SAPI -10..10 (TS4 Access default 3) onto AVSpeech min..max around default.
    let clamped = max(-10, min(10, r))
    let base = AVSpeechUtteranceDefaultSpeechRate
    let span = clamped >= 0 ? AVSpeechUtteranceMaximumSpeechRate - base
                            : base - AVSpeechUtteranceMinimumSpeechRate
    speaker.rate = base + span * clamped / 10 * 0.6
}
if let v = option("--volume").flatMap(Float.init) {
    speaker.volume = max(0, min(100, v)) / 100
}

let stdinThread = Thread {
    while let line = readLine(strippingNewline: true) {
        DispatchQueue.main.async { speaker.handle(line) }
    }
    // Game closed the pipe (quit or crash): stop and leave.
    DispatchQueue.main.async {
        speaker.synth.stopSpeaking(at: .immediate)
        exit(0)
    }
}
stdinThread.start()
RunLoop.main.run()
