// TS4 Access macOS speech companion.
//
// Drop-in replacement for the Windows SAPI child process that TS4 Access
// spawns from inside the game. Same stdin protocol, one command per line:
//   I<url-encoded text>   stop current speech, then speak text
//   Q<url-encoded text>   speak text after whatever is already queued
//   I                     stop speaking (used by the mod's stop key)
// The process exits when stdin closes, so it never outlives the game.
//
// With --hotkey (from the bridge): system shortcuts that work while the game
// has focus, Option+Command+S (speech on/off) and Option+Command+A (TS4
// Access menu), and a menu bar icon with the same actions plus the phone
// companion. Each prints "K voice", "K hub" or "K phone" on stdout. System
// hot keys need no Accessibility permission.
//
// Options: --voice <name or identifier substring>  --lang <bcp47, e.g. fr-FR>
//          --rate <-10..10, SAPI scale>  --volume <0..100>  --list-voices  --hotkey
//
// MIT licence, like TS4 Access.
import AVFoundation
import AppKit
import Carbon.HIToolbox
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

// Shortcuts and menu bar icon (only when the bridge asks, with --hotkey).
// Each prints a "K <name>" line on stdout for the bridge.
func report(_ name: String) {
    print("K \(name)")
    fflush(stdout)
}

// Hot keys are physical keys: find the key that types this letter on the
// current layout (on a French AZERTY keyboard "A" is where QWERTY has "Q").
func keyCode(typing letter: Character, fallback: Int) -> Int {
    guard let source = TISCopyCurrentKeyboardLayoutInputSource()?.takeRetainedValue(),
          let raw = TISGetInputSourceProperty(source, kTISPropertyUnicodeKeyLayoutData) else { return fallback }
    let data = Unmanaged<CFData>.fromOpaque(raw).takeUnretainedValue() as Data
    return data.withUnsafeBytes { bytes -> Int in
        guard let layout = bytes.baseAddress?.assumingMemoryBound(to: UCKeyboardLayout.self) else { return fallback }
        for code in 0..<128 {
            var dead: UInt32 = 0
            var chars = [UniChar](repeating: 0, count: 4)
            var length = 0
            let status = UCKeyTranslate(layout, UInt16(code), UInt16(kUCKeyActionDown), 0, UInt32(LMGetKbdType()),
                                        OptionBits(kUCKeyTranslateNoDeadKeysBit), &dead, chars.count, &length, &chars)
            if status == noErr, length == 1, Character(UnicodeScalar(chars[0])!).lowercased() == String(letter) {
                return code
            }
        }
        return fallback
    }
}

final class Controls: NSObject {
    static let keys: [(code: Int, name: String)] = [
        (keyCode(typing: "s", fallback: kVK_ANSI_S), "voice"),
        (keyCode(typing: "a", fallback: kVK_ANSI_A), "hub"),
    ]
    var hotKeys: [EventHotKeyRef?] = []
    var item: NSStatusItem?

    func install() {
        var pressed = EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))
        InstallEventHandler(GetApplicationEventTarget(), { _, event, _ in
            var id = EventHotKeyID()
            GetEventParameter(event, EventParamName(kEventParamDirectObject), EventParamType(typeEventHotKeyID),
                              nil, MemoryLayout<EventHotKeyID>.size, nil, &id)
            let index = Int(id.id)
            if index < Controls.keys.count { report(Controls.keys[index].name) }
            return noErr
        }, 1, &pressed, nil, nil)
        for (index, key) in Controls.keys.enumerated() {
            var ref: EventHotKeyRef?
            let status = RegisterEventHotKey(UInt32(key.code), UInt32(optionKey | cmdKey),
                                             EventHotKeyID(signature: 0x5453_4141, id: UInt32(index)), // "TSAA"
                                             GetApplicationEventTarget(), 0, &ref)
            if status != noErr {
                FileHandle.standardError.write("hot key \(key.name) unavailable (\(status))\n".data(using: .utf8)!)
            }
            hotKeys.append(ref)
        }

        let item = NSStatusBar.system.statusItem(withLength: NSStatusItem.squareLength)
        if let button = item.button {
            if let image = NSImage(systemSymbolName: "accessibility", accessibilityDescription: "TS4 Access") {
                image.isTemplate = true
                button.image = image
            } else {
                button.title = "TS4"
            }
            button.toolTip = "TS4 Access"
        }
        let menu = NSMenu()
        menu.addItem(entry("TS4 Access menu", key: "a", action: #selector(hub)))
        menu.addItem(entry("Speech on or off", key: "s", action: #selector(voice)))
        menu.addItem(.separator())
        menu.addItem(entry("Phone companion…", key: "", action: #selector(phone)))
        item.menu = menu
        self.item = item
    }

    func entry(_ title: String, key: String, action: Selector) -> NSMenuItem {
        let entry = NSMenuItem(title: title, action: action, keyEquivalent: key)
        entry.keyEquivalentModifierMask = [.option, .command]
        entry.target = self
        return entry
    }

    @objc func hub() { report("hub") }
    @objc func voice() { report("voice") }
    @objc func phone() { report("phone") }
}

let app = NSApplication.shared
let controls = Controls()
if CommandLine.arguments.contains("--hotkey") {
    // Menu bar icon, no Dock icon.
    app.setActivationPolicy(.accessory)
    controls.install()
} else {
    app.setActivationPolicy(.prohibited)
}
app.run()
