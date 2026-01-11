// jest.setup.ts
import '@testing-library/jest-dom/extend-expect';

// Mock the apiClient used in stores
jest.mock('../src/lib/axios', () => {
    const mockAxios = {
        get: jest.fn(),
        post: jest.fn(),
        defaults: { headers: {} },
    };
    return mockAxios;
});
