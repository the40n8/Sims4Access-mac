# Test checklist (Mac and phone app)

For a play session: try each line, and note what happened when it does not
match. Lines marked **new** have only been checked outside the game so far.
Install the latest build first (`mac/install.sh`, game closed).

## How to report a problem

For each problem, note:
1. **Where**: the screen (Live, Play, House, Sims, Settings, Help) or the Mac.
2. **What you did**: the buttons or keys, in order.
3. **What happened**: what VoiceOver or the Mac said, word for word if you can.
4. **What you expected**.
5. **When**: the game time or a rough clock time (it helps find the log line).

Example: "Play > Fridge > Cook. The question 'What should Bella cook?' did
not appear on the phone; the Mac showed it. 21:40."

## Starting

- [ ] Game loads a household; the Mac says "Phone link ready. Press Option
      Command A…" once you are in the house (not at the main menu). **new**
- [ ] Option+Command+A opens the TS4 Access menu, **phone companion** first,
      then the new grouped menu (yourself, needs, Sims, places…).
- [ ] Phone companion shows the QR code; the phone opens the app.
- [ ] Option+Command+S turns the Mac voice off and on ("Speech off" / "Speech on").

## Top bar and tabs

- [ ] First on screen: Pause (or Resume) with the speed and the time; it
      pauses and resumes at the previous speed. VoiceOver stays on the button.
- [ ] Six tabs in two rows; Help opens; long text does not run off the screen.

## Live: Needs attention **new**

- [ ] A low need appears with buttons (Eat now, Cook…, Sleep, Shower…).
- [ ] **Eat now** queues a quick meal from the fridge; **Cook…** opens the
      recipe question on the phone.
- [ ] A broken object, a puddle, spoiled food or food ready shows up, and its
      button works (Repair, Mop it up, Throw it away, Eat).
- [ ] A visitor on the lot: Greet / Welcome work.
- [ ] Bills due: **Pay the bills** pays them at the mailbox.
- [ ] A Sim you do not control shows "Control <name>" first.
- [ ] Each Sim line says which room the Sim is in.

## Game questions on the phone **new**

- [ ] A question with buttons (OK / Cancel or custom) appears on every tab
      as "The game asks: …" and is announced; answering it works in the game.
- [ ] A list to pick from (recipe, guests, food to order): rows read well,
      greyed rows say why, OK works only after choosing.
- [ ] A text box (naming something): typing and OK work.
- [ ] Corner notifications are announced and listed in History.

## Play

- [ ] Order: Controlled Sim, the Sim's actions (Cancel works), Camera, You
      (yourself and the phone), Interact, Game speed.
- [ ] Search "frigo" / "nevera" / "fridge"; the keyboard's Search key closes
      the keyboard and goes to the results.
- [ ] Categories open and close; items are inside.
- [ ] An object's menu lists the game's actions (not "No actions");
      choosing one queues it.
- [ ] The Sim's phone menu: call someone, invite over, order food. **new**
- [ ] Camera: Follow, Close-up, Free camera move the Mac's camera. **new**
- [ ] Settings > "Pause while I choose an action": pauses on opening a menu,
      resumes after queuing.

## House **new**

- [ ] Rooms are named sensibly (kitchen, bathroom, bedroom…), with size and
      where they are on the floor.
- [ ] Doors and openings link the right rooms; tapping one opens that room.
- [ ] Stairs: each floor says where they are; rooms list "Stairs up/down".
- [ ] An open kitchen and living room is named by both and says which part is where.
- [ ] Garden and pool appear with what is in them.

## Sims **new**

- [ ] Work performance, skills, relationships (with labels) look right.
- [ ] Goals: aspiration and milestone steps, wants and fears, work tasks,
      promotion steps.

## Conversations and announcements **new**

- [ ] A chat between Sims gives "Bella and Nancy: Chat, friendship +3."
- [ ] A new relationship level gives "now Friends" (or similar).
- [ ] Settings > Conversations changes how much is announced.
- [ ] Settings > The Mac voice announces: ticked kinds are said by the Mac,
      not when the Mac voice is off.

## Mac keyboard **new**

Only while the game is the front app (see KEYBOARD.md):
- [ ] Control+Shift+H (menu), +C (time), +Q (queue), +N (needs), +1/2 (switch Sims).
- [ ] Option+Shift+T (time), +N (needs), +R (rooms), +W (where am I).
- [ ] With keyboard navigation on (TS4 Access settings): arrows, Tab, Enter,
      Esc move through an open menu or dialog with speech; outside menus the
      game keeps its keys.
- [ ] Other apps keep their shortcuts while the game is in the background.

## Ask an AI

- [ ] Help > Ask an AI > Open ChatGPT: the request is filled in and the
      answer follows the guide in your language.
