/** @type {import('next').NextConfig} */
const getApiUrl = () => {
    if (process.env.API_URL) {
        return process.env.API_URL.startsWith('http')
            ? process.env.API_URL
            : `http://${process.env.API_URL}`
    }
    return 'http://localhost:8080'
}

const nextConfig = {
    reactStrictMode: true,
    output: 'standalone',
    async rewrites() {
        return [
            {
                source: '/api/:path*',
                destination: `${getApiUrl()}/api/:path*`,
            },
        ]
    },
}

module.exports = nextConfig
