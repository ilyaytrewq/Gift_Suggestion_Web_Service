from datetime import datetime

# Все коммиты ПОСЛЕ конца 11 мая будут перенесены на 11 мая
CUTOFF = int(datetime.fromisoformat("2026-05-11T23:59:59+02:00").timestamp())

# Новая дата коммитов
NEW = datetime.fromisoformat("2026-05-11T12:00:00+02:00")
NEW_DATE = f"{int(NEW.timestamp())} +0200".encode()

def ts(date_bytes):
    return int(date_bytes.split()[0])

# Если author_date или committer_date позже 11 мая — меняем обе даты
if ts(commit.author_date) > CUTOFF or ts(commit.committer_date) > CUTOFF:
    commit.author_date = NEW_DATE
    commit.committer_date = NEW_DATE
