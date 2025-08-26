function test() {
    return {
        name: 111,
        zzz: 222,
    }
}
//
// let x = new test()
// console.log(x.name)

for (const x in test()) {
    console.log(x)
}

// console.log(+3)
