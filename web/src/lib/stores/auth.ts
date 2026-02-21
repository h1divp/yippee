import { writable } from 'svelte/store';

export interface User {
	username: string;
	avatar?: string;
}

// 1. The Stores
export const user = writable<User | null>(null);
export const loading = writable(true); // Start true to block UI while we "check"

// Helper to simulate network delay (optional, but makes it feel real)
const delay = (ms: number) => new Promise((res) => setTimeout(res, ms));

// 2. Mock Login
export async function login(username: string, password: string) {
    // Simulate a failed login if password is "fail"
	if (password === 'fail') return false;

	// A. Create the fake user object
	const fakeUser: User = {
		username,
		avatar: `https://ui-avatars.com/api/?name=${username}&background=random`
	};

	localStorage.setItem('yippee_user', JSON.stringify(fakeUser));
	user.set(fakeUser);

	return true;
}

export async function checkSession() {
	loading.set(true);

	try {
		const stored = localStorage.getItem('yippee_user');

		if (stored) {
			const parsedUser = JSON.parse(stored);
			user.set(parsedUser);
		} else {
			user.set(null);
		}
	} catch (e) {
		console.error('Failed to parse session', e);
		user.set(null);
	} finally {
		loading.set(false);
	}
}

export async function logout() {
	localStorage.removeItem('yippee_user');
	user.set(null);
}
