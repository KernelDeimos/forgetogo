import { Form, Field, Formik } from "formik";

export class DyformikForm {
    constructor (fields) {
        this.fields = fields.map(v => this.normalizeField(v));
    }

    getFormik() {
        return (
            <Formik
                initialValues={this.fieldsToInitials(this.fields)}
                onSubmit={this.onSubmit}
            >
                {({ isSubmitting }) =>
                    this.createForm(isSubmitting)}
            </Formik>
        );
    }

    // === internal use below here ===
    createForm (isSubmitting) { return (
        <Form>
            {
                this.fields.map(v => this.fieldToElem(v)).flat()
            }
            <div></div>
            <button type="submit" disabled={isSubmitting}>
                Submit
            </button>
        </Form>
    );}

    // The following are really functions, not methods
    normalizeField (field) {
        let newField = { ...field };
        if ( ! newField.type ) {
            newField.type = 'text';
        }
        return newField;
    }
    fieldToInitial (field) {
        const defaultInitials = {
            'text': '',
            'hidden': '',
            'integer': 1,
            'checkbox': true,
        };
        let v = defaultInitials[field.type];
        if ( v === undefined ) v = '';
        if ( field.initial !== undefined ) {
            v = field.initial;
        }
        return v
    }
    fieldsToInitials (fields) {
        let o = {};
        for ( let field of fields ) {
            o[field.name] = this.fieldToInitial(field);
        }
        return o
    }
    fieldToElem (field) {
        let elem;
        if ( field.type === 'integer' ) {
            elem = (
                <Field
                    name={field.name}
                    type="number"
                    min="1"
                    step="1"
                    pattern="\d*"
                ></Field>
            );
        } else if ( field.type === 'hidden' ) {
            elem = [];
        } else {
            elem = (
                <Field
                    name={field.name}
                    type={field.type}
                ></Field>
            )
        }
        return [
            (
                <label htmlFor={field.name}>{field.label || field.name}</label>
            ), elem
        ];
    }
}