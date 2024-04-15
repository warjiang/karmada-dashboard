import './App.css'
import Router from "./routes";
import {Helmet, HelmetProvider} from 'react-helmet-async';
import {ConfigProvider} from 'antd';

function App() {
    return (
        <ConfigProvider
            theme={{
                components: {
                    Layout: {
                        siderBg: '#ffffff'
                    }
                }
            }}
        >
            <HelmetProvider>
                <Helmet>
                    <title>Karmada Dashboard</title>
                    <link rel="apple-touch-icon" sizes="180x180" href="/public/apple-touch-icon.png"/>
                    <link rel="icon" type="image/png" sizes="16x16" href="/public/favicon-16x16.png"/>
                    <link rel="icon" type="image/png" sizes="32x32" href="/public/favicon-32x32.png"/>
                    <link rel="shortcut icon" type="image/x-icon" href="/public/favicon.ico"/>
                </Helmet>
                <Router/>
            </HelmetProvider>
        </ConfigProvider>
    )
}

export default App
