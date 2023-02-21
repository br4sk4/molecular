import { HopeProvider } from "@hope-ui/solid"
import Application from "./Application"

window.addEventListener("contextmenu", e => e.preventDefault());
window.addEventListener("contextmenu", e => e.preventDefault());

function ApplicationProvider() {
    const config = {
        initialColorMode: "dark",
        darkTheme: {
            colors: {
                background: "#222222",
                color: "white",
                primary1: "#330d43",
                primary2: "#3d1051",
                primary3: "#52156c",
                primary4: "#5c177a",
                primary5: "#661a87",
                primary6: "#711c94",
                primary7: "#8c23b8",
                primary8: "#9d27ce",
                primary9: "#a937d9",
                primary10: "#b24edd",
                primary11: "#bc64e1",
                primary12: "#c67ae6",
            },
        },
    }

    return (
        <HopeProvider config={config}>
            <Application />
        </HopeProvider>
    )
}

export default ApplicationProvider