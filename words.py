import os
import random
import re
import sys


def normalize_answer(text):
    return re.sub(r"[^a-z0-9]+", " ", text.casefold()).strip()


def parse_line(line):
    parts = [part.strip() for part in line.split("=>")]
    term = parts[0]
    translation_text = parts[1] if len(parts) > 1 else ""
    description = "=>".join(parts[2:]).strip() if len(parts) > 2 else ""

    translations = [alt.strip() for alt in translation_text.split("/") if alt.strip()]

    if not translations and translation_text:
        translations = [translation_text.strip()]

    return {
        "term": term,
        "translation_text": translation_text,
        "translations": translations,
        "description": description,
    }


def shuffle_array(arr):
    shuffled = arr[:]
    random.shuffle(shuffled)
    return shuffled


def get_filename():
    filename = sys.argv[1] if len(sys.argv) > 1 else None

    if filename is None:
        files = [
            file.replace(".txt", "")
            for file in os.listdir("exercises")
            if file.endswith(".txt")
        ]
        print("Available files:")
        for fileIndex, file in enumerate(files):
            print(f"- {fileIndex + 1}. {file}")
        filename = input("Enter the name or the number of the file to use: ")
        if filename.isdigit():
            filename = files[(int(filename) % len(files)) - 1]
        if filename not in files:
            print(f"File '{filename}' not found. Using first file instead.")
            filename = files[0]
    return f"exercises/{filename}.txt"


def ignore_comments(lines):
    return [line for line in lines if line.strip() and not line.strip().startswith("#")]


def main():
    filename = get_filename()

    with open(filename, "r", encoding="utf-8") as file:
        lines = ignore_comments(file.readlines())

    original_entries = [parse_line(line) for line in lines]
    active_entries = original_entries[:]
    skipped = set()

    round_num = 1
    while active_entries:
        print(f"\n--- Round {round_num} ---\n")
        active_entries = shuffle_array(active_entries)
        remaining = []

        for i, entry in enumerate(active_entries):
            term = entry["term"]
            translations = entry["translations"]
            translation_text = entry["translation_text"]
            description = entry["description"]

            print(f"\n[{i + 1}/{len(active_entries)}] 🇫🇷 {term}")
            user_input = (
                input("(Guess the translation or press ↵ to reveal) ").strip().lower()
            )

            valid_answers = {normalize_answer(alt) for alt in translations}
            if translation_text.strip():
                valid_answers.add(normalize_answer(translation_text))
            normalized_input = normalize_answer(user_input)
            skipped_this_round = False

            if user_input and user_input == "s":
                user_input = ""

            if user_input:
                if normalized_input in valid_answers:
                    print("✅ Correct guess!")
                    skipped.add(term)
                    skipped_this_round = True
                else:
                    print("❌ Not quite.")
            else:
                print("...")

            purple = "\033[95m"
            reset = "\033[0m"
            translation_display = (
                translation_text if translation_text else "[No translation]"
            )
            line_output = (
                f"[{i + 1}/{len(active_entries)}] 🇫🇷 {term} => 🇬🇧 {translation_display}"
            )
            if description:
                line_output += f" => {purple}{description}{reset}"

            print(line_output)
            if not skipped_this_round:
                follow_up = (
                    input(
                        "Did you know it? Press 's' to skip next time, ↵ to keep reviewing: "
                    )
                    .strip()
                    .lower()
                )
                if follow_up == "s":
                    skipped.add(term)

            remaining.append(entry)

        # Filter out skipped items
        active_entries = [entry for entry in remaining if entry["term"] not in skipped]
        round_num += 1

    print("\n✅ All words reviewed or skipped. Goodbye!")


if __name__ == "__main__":
    main()
