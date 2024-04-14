import {createBrowserRouter, redirect, RouterProvider} from "react-router-dom";
import {MainLayout} from '@/layout'
import ErrorBoundary from '@/components/error'
import Overview from '@/pages/overview'
import Login from '@/pages/login';

const redirectToHomepage = () => {
    return redirect("/overview");
};

const router = createBrowserRouter([
        {
            path: "/",
            element: <MainLayout/>,
            errorElement: <ErrorBoundary/>,
            children: [
                {
                    path: "/",
                    loader: redirectToHomepage,
                },
                {
                    path: "/overview",
                    element: <Overview/>,
                }
            ],
        },
        {
            path: "/login",
            errorElement: <ErrorBoundary/>,
            element: <Login/>
        }
    ],
);

export default function Router() {
    return <RouterProvider router={router}/>;
}
