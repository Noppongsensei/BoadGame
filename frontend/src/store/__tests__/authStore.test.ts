const mockPost = jest.fn();
jest.mock('../../lib/axios', () => ({
    __esModule: true,
    default: {
        post: mockPost,
        get: jest.fn(),
        defaults: { headers: {} }
    }
}));

const { useAuthStore } = require('../authStore');
const apiClient = require('../../lib/axios').default;

describe('authStore', () => {
    beforeEach(() => {
        // reset store state
        const { reset } = useAuthStore.getState();
        // Zustand doesn't have reset by default, we recreate store
        // Simple approach: re-import store to get fresh state
    });

    it('login success stores token and user', async () => {
        const mockResponse = { data: { user: { id: '1', username: 'test' }, token: 'jwt-token' } };
        // @ts-ignore
        apiClient.post.mockResolvedValueOnce(mockResponse);

        const { login, token, user, isAuthenticated } = useAuthStore.getState();
        await login('test', 'pass');

        expect(apiClient.post).toHaveBeenCalledWith('/api/auth/login', { username: 'test', password: 'pass' });
        const state = useAuthStore.getState();
        expect(state.token).toBe('jwt-token');
        expect(state.user?.username).toBe('test');
        expect(state.isAuthenticated).toBe(true);
    });

    it('register failure sets error', async () => {
        // @ts-ignore
        apiClient.post.mockRejectedValueOnce({ response: { data: { error: 'User exists' } } });
        const { register } = useAuthStore.getState();
        await expect(register('test', 'pass')).rejects.toBeDefined();
        const state = useAuthStore.getState();
        expect(state.error).toBe('User exists');
    });
});
