type SessionListener = () => void;

let accessToken: string | null = null;

const listeners = new Set<SessionListener>();

function notifyListeners(): void {
  listeners.forEach((listener) => listener());
}

export const authSession = {
  getAccessToken(): string | null {
    return accessToken;
  },
  setAccessToken(nextToken: string): void {
    accessToken = nextToken;
    notifyListeners();
  },
  clearAccessToken(): void {
    accessToken = null;
    notifyListeners();
  },
  subscribe(listener: SessionListener): () => void {
    listeners.add(listener);

    return () => {
      listeners.delete(listener);
    };
  },
};
