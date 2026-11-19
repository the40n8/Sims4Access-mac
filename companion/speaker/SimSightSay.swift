// SimSight macOS speech companion.
//
// Drop-in replacement for the Windows SAPI child process that SimSight
// spawns from inside the game. Same stdin protocol, one command per line:
//   I<url-encoded text>   stop current speech, then speak text
//   Q<url-encoded text>   speak text after whatever is already queued
//   I                     stop speaking (used by the mod's stop key)
// The process exits when stdin closes, so it never outlives the game.
//
// With --hotkey (from the bridge): system shortcuts that work while the game
// has focus, Option+Command+S (speech on/off) and Option+Command+A (TS4
// Access menu), and a menu bar icon with the same actions plus the phone
// companion. Each prints "K voice", "K hub" or "K phone" on stdout.
// While The Sims 4 is the front app it also takes the original mod's
// keyboard (Control+Shift for its Ctrl+Alt keys, Option+Shift for its Alt
// keys, and the bare keys the game says the mod owns, sent as "G vk,..."
// lines on stdin) and prints "K key <family> <vk> <down|up> <shift>".
// System hot keys need no Accessibility permission.
//
// Options: --voice <name or identifier substring>  --lang <bcp47, e.g. fr-FR>
//          --rate <-10..10, SAPI scale>  --volume <0..100>  --list-voices  --hotkey
//
// MIT licence, like SimSight.
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
        if kind == "G" {
            // Bare keys the mod owns now (from the game, through the bridge).
            Controls.shared?.setGrabs(String(line.dropFirst()))
            return
        }
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
    // SAPI -10..10 (SimSight default 3) onto AVSpeech min..max around default.
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

/// One registered hot key: Mac key code + modifiers, reported as
/// "K <name>" (global ones) or "K key <family> <Windows VK> <down|up> <shift>".
struct HotKey {
    let code: Int
    let mods: Int
    let name: String     // global: "voice", "hub"
    let family: String   // game keys: chord, needs, plus, plain
    let vk: Int
    let shift: Bool
}

// Windows VK codes of the keys the original mod uses, and the Mac key that
// types them: letters follow the keyboard layout, the rest are positions.
func macKey(forVK vk: Int) -> Int? {
    switch vk {
    case 0x41...0x5A:
        let letter = Character(UnicodeScalar(UInt8(vk + 32)))
        return keyCode(typing: letter, fallback: -1) >= 0 ? keyCode(typing: letter, fallback: -1) : nil
    case 0x31: return kVK_ANSI_1
    case 0x32: return kVK_ANSI_2
    case 0x33: return kVK_ANSI_3
    case 0x34: return kVK_ANSI_4
    case 0x35: return kVK_ANSI_5
    case 0x36: return kVK_ANSI_6
    case 0x37: return kVK_ANSI_7
    case 0x38: return kVK_ANSI_8
    case 0x25: return kVK_LeftArrow
    case 0x26: return kVK_UpArrow
    case 0x27: return kVK_RightArrow
    case 0x28: return kVK_DownArrow
    case 0x21: return kVK_PageUp
    case 0x22: return kVK_PageDown
    case 0x24: return kVK_Home
    case 0x2E: return kVK_ForwardDelete
    case 0x0D: return kVK_Return
    case 0x20: return kVK_Space
    case 0x08: return kVK_Delete
    case 0x09: return kVK_Tab
    case 0x1B: return kVK_Escape
    case 0xBC: return kVK_ANSI_Comma
    case 0xBE: return kVK_ANSI_Period
    case 0xDB: return kVK_ANSI_LeftBracket
    case 0xDD: return kVK_ANSI_RightBracket
    default: return nil
    }
}

func vk(_ c: Character) -> Int { Int(c.asciiValue!) }

// The original's Ctrl+Alt+key chords (Mac: Control+Shift+key).
let chordVKs: [Int] = [0x26, 0x28, 0x25, 0x27, 0x21, 0x22, 0x0D, 0x20, 0x08, 0xBC, 0xBE, 0xDB, 0xDD]
    + "SWRHCPEGQLVMKOIUDJBTZYXF".map(vk) + Array(0x31...0x38)
// The original's Alt+key Plus shortcuts (Mac: Option+Shift+key; with Command
// for the Shift variant of the drawing-cursor arrows).
let plusVKs: [Int] = "AQNLTRWCBXSZD".map(vk) + [0x25, 0x26, 0x27, 0x28]

final class Controls: NSObject {
    static var shared: Controls?
    // SIMSIGHT_GAME_BUNDLE: another app counts as the game (for testing).
    static let gameBundle = ProcessInfo.processInfo.environment["SIMSIGHT_GAME_BUNDLE"] ?? "com.ea.mac.thesims4"

    var table: [UInt32: HotKey] = [:]
    var global: [EventHotKeyRef] = []
    var game: [EventHotKeyRef] = []
    var grabs = ""
    var gameFront = false
    var item: NSStatusItem?

    func install() {
        Controls.shared = self
        var events = [EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed)),
                      EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyReleased))]
        InstallEventHandler(GetApplicationEventTarget(), { _, event, _ in
            var id = EventHotKeyID()
            GetEventParameter(event, EventParamName(kEventParamDirectObject), EventParamType(typeEventHotKeyID),
                              nil, MemoryLayout<EventHotKeyID>.size, nil, &id)
            let down = GetEventKind(event) == UInt32(kEventHotKeyPressed)
            Controls.shared?.fired(id.id, down: down)
            return noErr
        }, 2, &events, nil, nil)

        // Always: speech on/off and the SimSight menu.
        for (name, letter) in [("voice", Character("s")), ("hub", Character("a"))] {
            let code = keyCode(typing: letter, fallback: letter == "s" ? kVK_ANSI_S : kVK_ANSI_A)
            if let ref = register(HotKey(code: code, mods: optionKey | cmdKey, name: name, family: "", vk: 0, shift: false)) {
                global.append(ref)
            }
        }

        // Game keys: only while The Sims 4 is the front app.
        let center = NSWorkspace.shared.notificationCenter
        center.addObserver(self, selector: #selector(activated(_:)),
                           name: NSWorkspace.didActivateApplicationNotification, object: nil)
        updateFront(NSWorkspace.shared.frontmostApplication)
        installMenu()
    }

    var nextID: UInt32 = 1
    func register(_ key: HotKey) -> EventHotKeyRef? {
        var ref: EventHotKeyRef?
        let id = nextID
        nextID += 1
        let status = RegisterEventHotKey(UInt32(key.code), UInt32(key.mods), EventHotKeyID(signature: 0x5453_4141, id: id),
                                         GetApplicationEventTarget(), 0, &ref)
        if status != noErr {
            FileHandle.standardError.write("hot key \(key.name)\(key.family) \(key.vk) unavailable (\(status))\n".data(using: .utf8)!)
            return nil
        }
        table[id] = key
        return ref
    }

    func fired(_ id: UInt32, down: Bool) {
        guard let key = table[id] else { return }
        if key.family.isEmpty {
            if down { report(key.name) }
            return
        }
        report("key \(key.family) \(key.vk) \(down ? "down" : "up") \(key.shift ? 1 : 0)")
    }

    @objc func activated(_ note: Notification) {
        updateFront(note.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication)
    }

    func updateFront(_ app: NSRunningApplication?) {
        let front = app?.bundleIdentifier == Controls.gameBundle
        if front == gameFront { return }
        gameFront = front
        if front {
            registerGameKeys()
        } else {
            unregisterGameKeys()
            report("key focus 0 up 0")   // release anything held
        }
    }

    func registerGameKeys() {
        unregisterGameKeys()
        let chord = controlKey | shiftKey
        for v in chordVKs {
            if let code = macKey(forVK: v),
               let ref = register(HotKey(code: code, mods: chord, name: "", family: "chord", vk: v, shift: false)) {
                game.append(ref)
            }
        }
        if let code = macKey(forVK: vk("N")),
           let ref = register(HotKey(code: code, mods: chord, name: "", family: "needs", vk: vk("N"), shift: false)) {
            game.append(ref)
        }
        for v in plusVKs {
            guard let code = macKey(forVK: v) else { continue }
            if let ref = register(HotKey(code: code, mods: optionKey | shiftKey, name: "", family: "plus", vk: v, shift: false)) {
                game.append(ref)
            }
            if (0x25...0x28).contains(v),
               let ref = register(HotKey(code: code, mods: optionKey | shiftKey | cmdKey, name: "", family: "plus", vk: v, shift: true)) {
                game.append(ref)
            }
        }
        for item in grabs.split(separator: ",") {
            let parts = item.split(separator: "+")
            guard let v = Int(parts[0]), let code = macKey(forVK: v) else { continue }
            let shifted = parts.count > 1
            if let ref = register(HotKey(code: code, mods: shifted ? shiftKey : 0, name: "", family: "plain", vk: v, shift: shifted)) {
                game.append(ref)
            }
        }
    }

    func unregisterGameKeys() {
        for ref in game { UnregisterEventHotKey(ref) }
        game.removeAll()
        table = table.filter { $0.value.family.isEmpty }
    }

    /// "vk,vk+shift,...": the bare keys the mod owns; taken only while the game is in front.
    func setGrabs(_ list: String) {
        let trimmed = list.trimmingCharacters(in: .whitespaces)
        if trimmed == grabs { return }
        grabs = trimmed
        if gameFront { registerGameKeys() }
    }

    func installMenu() {
        let item = NSStatusBar.system.statusItem(withLength: NSStatusItem.squareLength)
        if let button = item.button {
            if let image = NSImage(systemSymbolName: "accessibility", accessibilityDescription: "SimSight") {
                image.isTemplate = true
                button.image = image
            } else {
                button.title = "TS4"
            }
            button.toolTip = "SimSight"
        }
        let menu = NSMenu()
        menu.addItem(entry("SimSight menu", key: "a", action: #selector(hub)))
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
