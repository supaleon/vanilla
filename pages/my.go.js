export const UserTag = () => ({
  Id: 0,
  Name: "",
});

export const Location = () => ({
  Street: "",
  ZipCode: 0,
});

export const User = () => ({
  Name: "",
  Location: Location(),
  Tags: [UserTag()],
  Followers: 0,
  Hello:{}
});

alert(User().Hello.abc.d)


/**
 * @template T
@param {NonNullable<Exclude<T, Function>>} Type
@returns {NonNullable<Exclude<T, Function>>}
 */
function props2(Type) {
    return Type
}


let any = {};







