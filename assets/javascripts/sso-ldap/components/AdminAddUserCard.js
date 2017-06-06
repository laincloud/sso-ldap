import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminAddUserCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'name': true,
        'email': true,
        'password': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">添加用户</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onAdd, [
              this.renderInput("name", "用户名*(字母、数字和减号)", { type: "text", pattern: "[\-a-zA-Z0-9]*" }),
              this.renderInput("fullname", "全名", { type: 'text' }),
              this.renderInput("email", "Email*(仅限公司Email地址)", { type: 'email' }),
              this.renderInput("password", "密码*", { type: 'password' }),
              this.renderInput("mobile", "手机号", { type: 'tel' }),
            ])
        }
        { this.renderAction("确定", this.onAdd) }
      </div>
    );
  },

  onAdd() {
    const fields = ['name', 'fullname', 'email', 'password', 'mobile'];
    const rFields = ['name', 'email', 'password'];
    const {isValid, formData} = this.validateForm(fields, rFields);
    if (isValid) {
      const {token, tokenType} = this.props;
      this.setState({ inRequest: true });
      Admin.addUser(token, tokenType, formData, this.onRequestCallback);
    }
  },

});

export default AdminAddUserCard;
