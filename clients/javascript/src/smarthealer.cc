#include <napi.h>
#include "libsmarthealer.h"
#include <string>

// Convert C string to Napi::String safely
Napi::String SafeCStringToNapi(Napi::Env env, const char* str) {
    return str ? Napi::String::New(env, str) : Napi::String::New(env, "");
}

// Convert Result struct to JavaScript object
Napi::Object ResultToJS(Napi::Env env, const Result& result) {
    Napi::Object obj = Napi::Object::New(env);
    obj.Set("success", Napi::Boolean::New(env, result.success != 0));
    obj.Set("reason", SafeCStringToNapi(env, result.reason));
    obj.Set("content", SafeCStringToNapi(env, result.content));
    return obj;
}

// Convert JavaScript object to Config struct
Config JSToConfig(const Napi::Object& jsConfig) {
    Config config = {nullptr, nullptr};

    if (jsConfig.Has("openai_key")) {
        std::string openaiKey = jsConfig.Get("openai_key").As<Napi::String>().Utf8Value();
        config.openai_key = openaiKey.c_str();
    }

    if (jsConfig.Has("sqlite_db_path")) {
        std::string dbPath = jsConfig.Get("sqlite_db_path").As<Napi::String>().Utf8Value();
        config.sqlite_db_path = dbPath.c_str();
    }

    return config;
}

// Convert JavaScript object to Info struct
Info JSToInfo(const Napi::Object& jsInfo) {
    Info info = {nullptr, nullptr, nullptr, nullptr, nullptr, Android, XMLPageType};

    if (jsInfo.Has("project_id")) {
        std::string projectId = jsInfo.Get("project_id").As<Napi::String>().Utf8Value();
        info.project_id = projectId.c_str();
    }

    if (jsInfo.Has("page_source")) {
        std::string pageSource = jsInfo.Get("page_source").As<Napi::String>().Utf8Value();
        info.page_source = pageSource.c_str();
    }

    if (jsInfo.Has("b64_png")) {
        std::string b64Png = jsInfo.Get("b64_png").As<Napi::String>().Utf8Value();
        info.b64_png = b64Png.c_str();
    }

    if (jsInfo.Has("xpath")) {
        std::string xpath = jsInfo.Get("xpath").As<Napi::String>().Utf8Value();
        info.xpath = xpath.c_str();
    }

    if (jsInfo.Has("context_id")) {
        std::string contextId = jsInfo.Get("context_id").As<Napi::String>().Utf8Value();
        info.context_id = contextId.c_str();
    }

    if (jsInfo.Has("platform")) {
        info.platform = static_cast<Platform>(jsInfo.Get("platform").As<Napi::Number>().Int32Value());
    }

    if (jsInfo.Has("page_type")) {
        info.page_type = static_cast<PageType>(jsInfo.Get("page_type").As<Napi::Number>().Int32Value());
    }

    return info;
}

// Convert JavaScript object to Options struct
Options JSToOptions(const Napi::Object& jsOptions) {
    Options options = {Automatic};

    if (jsOptions.Has("comparisionMode")) {
        options.comparisionMode = static_cast<ComparisionMode>(
            jsOptions.Get("comparisionMode").As<Napi::Number>().Int32Value()
        );
    }

    return options;
}

// Wrapper for initSmartHealer
Napi::Value InitSmartHealer(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();

    // Validate arguments
    if (info.Length() < 1 || !info[0].IsObject()) {
        Napi::TypeError::New(env, "Expected config object as first argument")
            .ThrowAsJavaScriptException();
        return env.Undefined();
    }

    try {
        Napi::Object configObj = info[0].As<Napi::Object>();
        Config config = JSToConfig(configObj);

        Result result = initSmartHealer(config);
        Napi::Object jsResult = ResultToJS(env, result);

        // Clean up the result
        freeResult(result);

        return jsResult;
    } catch (const std::exception& e) {
        Napi::Error::New(env, e.what()).ThrowAsJavaScriptException();
        return env.Undefined();
    }
}

// Wrapper for resolveLocator
Napi::Value ResolveLocator(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();

    // Validate arguments
    if (info.Length() < 2 || !info[0].IsObject() || !info[1].IsObject()) {
        Napi::TypeError::New(env, "Expected info and options objects as arguments")
            .ThrowAsJavaScriptException();
        return env.Undefined();
    }

    try {
        Napi::Object infoObj = info[0].As<Napi::Object>();
        Napi::Object optionsObj = info[1].As<Napi::Object>();

        Info infoStruct = JSToInfo(infoObj);
        Options optionsStruct = JSToOptions(optionsObj);

        Result result = resolveLocator(infoStruct, optionsStruct);
        Napi::Object jsResult = ResultToJS(env, result);

        // Clean up the result
        freeResult(result);

        return jsResult;
    } catch (const std::exception& e) {
        Napi::Error::New(env, e.what()).ThrowAsJavaScriptException();
        return env.Undefined();
    }
}

// Wrapper for resolveLocatorAsync
Napi::Value ResolveLocatorAsync(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();

    // Validate arguments
    if (info.Length() < 2 || !info[0].IsObject() || !info[1].IsObject()) {
        Napi::TypeError::New(env, "Expected info and options objects as arguments")
            .ThrowAsJavaScriptException();
        return env.Undefined();
    }

    try {
        Napi::Object infoObj = info[0].As<Napi::Object>();
        Napi::Object optionsObj = info[1].As<Napi::Object>();

        Info infoStruct = JSToInfo(infoObj);
        Options optionsStruct = JSToOptions(optionsObj);

        Result result = resolveLocatorAsync(infoStruct, optionsStruct);
        Napi::Object jsResult = ResultToJS(env, result);

        // Clean up the result
        freeResult(result);

        return jsResult;
    } catch (const std::exception& e) {
        Napi::Error::New(env, e.what()).ThrowAsJavaScriptException();
        return env.Undefined();
    }
}

// Wrapper for close
Napi::Value Close(const Napi::CallbackInfo& info) {
    close();
    return info.Env().Undefined();
}

// Export constants for enums
Napi::Object CreateConstants(Napi::Env env) {
    Napi::Object constants = Napi::Object::New(env);

    // Platform enum
    Napi::Object platform = Napi::Object::New(env);
    platform.Set("Android", Napi::Number::New(env, static_cast<int>(Android)));
    platform.Set("Ios", Napi::Number::New(env, static_cast<int>(Ios)));
    platform.Set("Web", Napi::Number::New(env, static_cast<int>(Web)));
    constants.Set("Platform", platform);

    // PageType enum
    Napi::Object pageType = Napi::Object::New(env);
    pageType.Set("XML", Napi::Number::New(env, static_cast<int>(XMLPageType)));
    pageType.Set("HTML", Napi::Number::New(env, static_cast<int>(HTMLPageType)));
    constants.Set("PageType", pageType);

    // ComparisionMode enum
    Napi::Object comparisionMode = Napi::Object::New(env);
    comparisionMode.Set("Automatic", Napi::Number::New(env, static_cast<int>(Automatic)));
    comparisionMode.Set("Manual", Napi::Number::New(env, static_cast<int>(Manual)));
    comparisionMode.Set("Screenshot", Napi::Number::New(env, static_cast<int>(Screenshot)));
    constants.Set("ComparisionMode", comparisionMode);

    return constants;
}

// Module initialization
Napi::Object Init(Napi::Env env, Napi::Object exports) {
    // Export functions
    exports.Set("initSmartHealer", Napi::Function::New(env, InitSmartHealer));
    exports.Set("resolveLocator", Napi::Function::New(env, ResolveLocator));
    exports.Set("resolveLocatorAsync", Napi::Function::New(env, ResolveLocatorAsync));
    exports.Set("close", Napi::Function::New(env, Close));

    // Export constants
    exports.Set("constants", CreateConstants(env));

    return exports;
}

NODE_API_MODULE(smarthealer, Init)