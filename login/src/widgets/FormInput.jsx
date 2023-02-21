import {css, Input} from "@hope-ui/solid";

const formInputCSS = {
    color: "$primary",
    backgroundColor: "$neutral1",
    border: "1px solid $neutral7",
    "&:hover": {
        border: "1px solid $neutral10",
    },
    "&:focus": {
        boxShadow: "none",
        border: "1px solid $primary8",
    },
}

function FormInput(props) {
    return (
        <Input css={formInputCSS} {...props}/>
    )
}

export default FormInput